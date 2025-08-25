package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopkg.in/natefinch/lumberjack.v2"

	"skywalking_transformer/converter"
	_ "skywalking_transformer/docs"
	"skywalking_transformer/otel"
	"skywalking_transformer/skywalking"
)

// ----------- Config (env-driven) -----------
var (
	collectorURL    string
	receiverPort    string
	queueSize       int
	workerCount     int
	batchSize       int
	batchFlush      time.Duration
	httpTimeout     time.Duration
	shutdownTimeout time.Duration
	queueDropOnFull bool
)

// ----------- Helpers for env parsing -----------
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}
func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		switch v {
		case "1", "true", "TRUE", "True", "yes", "YES":
			return true
		case "0", "false", "FALSE", "False", "no", "NO":
			return false
		}
	}
	return def
}
func getenvDurMS(key string, defMS int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Millisecond
		}
	}
	return time.Duration(defMS) * time.Millisecond
}

// ----------- Logging -----------
func initLogger() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "./skywalking-transformer.log",
		MaxSize:    5, // MB
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	})
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// ----------- HTTP Client (reused) -----------
var httpClient *http.Client

func makeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}
}

// ----------- Async pipeline types -----------
type job struct {
	payload otel.OTelPayload
}
type combined struct {
	payload otel.OTelPayload
}

var (
	jobCh      chan job
	combinedCh chan combined
	wg         sync.WaitGroup
)

// @title SkyWalking Collector API Example
// @version 1.0
// @description This is an example API to receive SkyWalking agent payloads.
// @host localhost:8081
// @BasePath /
func main() {
	initLogger()
	log.Println("Starting CodeXray service...")

	// Core endpoints (env)
	collectorURL = os.Getenv("CODEXRAY_COLLECTOR_URL")
	if collectorURL == "" {
		collectorURL = "http://labs.codexray.io:8041/v1/traces"
	}
	receiverPort = os.Getenv("CODEXRAY_RECEIVER_PORT")
	if receiverPort == "" {
		receiverPort = "8081"
	}

	// Performance tuning (env)
	queueSize = getenvInt("CODEXRAY_QUEUE_SIZE", 50000)
	workerCount = getenvInt("CODEXRAY_WORKERS", runtime.NumCPU()*2)
	batchSize = getenvInt("CODEXRAY_BATCH_SIZE", 200)
	batchFlush = getenvDurMS("CODEXRAY_BATCH_FLUSH_MS", 100)
	httpTimeout = getenvDurMS("CODEXRAY_HTTP_TIMEOUT_MS", 5000)
	shutdownTimeout = getenvDurMS("CODEXRAY_SHUTDOWN_TIMEOUT_MS", 10000)
	queueDropOnFull = getenvBool("CODEXRAY_QUEUE_DROP_ON_FULL", false)

	log.Printf("Using OTEL collector endpoint: %s", collectorURL)
	log.Printf("Listening on port: %s", receiverPort)

	httpClient = makeHTTPClient(httpTimeout)

	// Build pipeline
	jobCh = make(chan job, queueSize)
	combinedCh = make(chan combined, workerCount*2)

	ctx, cancel := context.WithCancel(context.Background())

	// Batcher
	wg.Add(1)
	go func() {
		defer wg.Done()
		runBatcher(ctx)
	}()

	// Workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runSender(ctx, id)
		}(i + 1)
	}

	// HTTP server (Gin)
	r := gin.New()
	r.Use(gin.LoggerWithWriter(log.Writer()), gin.Recovery())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.POST("/v3/segments", collectAndEnqueueHandler)
	r.POST("/v3/management/reportProperties", reportPropertiesHandler)
	r.POST("/v3/management/keepAlive", keepAliveHandler)
	r.POST("/v3/clrMetricReports", clrMetricReportsHandler)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	srv := &http.Server{
		Addr:         ":" + receiverPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutdown signal received...")
		cancel()
		close(jobCh)
		ctxTimeout, cancel2 := context.WithTimeout(context.Background(), shutdownTimeout)
		if err := srv.Shutdown(ctxTimeout); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		defer cancel2()
		wg.Wait()
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// Call cancel directly instead of defer to avoid linter warning
		cancel()
		log.Fatalf("Server failed: %v", err)
	}
	<-idleConnsClosed
	log.Println("Shutdown complete.")
}

// ----------- Handlers -----------
func reportPropertiesHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "received"})
}

func collectAndEnqueueHandler(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to read body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var payload []skywalking.TraceSegment
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Bind error: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	enqueued := 0
	for i := range payload {
		segment := &payload[i] // Use pointer to avoid copying
		otelPayload := converter.SkywalkingToOtel(segment)
		j := job{payload: otelPayload}
		if queueDropOnFull {
			select {
			case jobCh <- j:
				enqueued++
			default:
				log.Printf("Queue full, dropping payload")
			}
		} else {
			jobCh <- j
			enqueued++
		}
	}

	c.JSON(200, gin.H{"status": "queued", "enqueued": enqueued})
}

func keepAliveHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "alive"})
}
func clrMetricReportsHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// ----------- Batcher & Sender -----------
func runBatcher(ctx context.Context) {
	ticker := time.NewTicker(batchFlush)
	defer ticker.Stop()
	var buf []otel.OTelPayload
	flush := func() {
		if len(buf) == 0 {
			return
		}
		select {
		case combinedCh <- combined{payload: mergePayloads(buf)}:
		case <-ctx.Done():
		}
		buf = buf[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			close(combinedCh)
			return
		case j, ok := <-jobCh:
			if !ok {
				flush()
				close(combinedCh)
				return
			}
			buf = append(buf, j.payload)
			if len(buf) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func runSender(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmb, ok := <-combinedCh:
			if !ok {
				return
			}
			if err := sendToCollector(cmb.payload); err != nil {
				log.Printf("[worker %d] send error: %v", id, err)
				// retry once
				if err2 := sendToCollector(cmb.payload); err2 != nil {
					log.Printf("[worker %d] retry error: %v", id, err2)
				}
			}
		}
	}
}

func mergePayloads(items []otel.OTelPayload) otel.OTelPayload {
	out := otel.OTelPayload{ResourceSpans: make([]otel.ResourceSpan, 0, len(items))}
	for _, it := range items {
		out.ResourceSpans = append(out.ResourceSpans, it.ResourceSpans...)
	}
	return out
}

func sendToCollector(p otel.OTelPayload) error {
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, collectorURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("otel backend returned status: %s", resp.Status)
	}
	return nil
}
