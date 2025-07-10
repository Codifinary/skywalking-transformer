package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sky2otel/converter"
	"sky2otel/otel"
	"sky2otel/skywalking"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "sky2otel/docs"
)

// @title SkyWalking Collector API Example
// @version 1.0
// @description This is an example API to receive SkyWalking agent payloads.
// @host localhost:8080
// @BasePath /
func main() {
	r := gin.Default()

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register POST endpoint for payload
	//r.POST("/collect", collectHandler)
	//r.POST("/collect-otel", collectOtelHandler)
	//r.POST("/convert-skywalking-to-otel", convertSkywalkingToOtelHandler)
	r.POST("/v3/segments", collectAndSendToOtelHandler)
	r.POST("/v3/management/reportProperties", reportPropertiesHandler)
	r.POST("/v3/management/keepAlive", keepAliveHandler)
	r.POST("/v3/clrMetricReports", clrMetricReportsHandler)

	r.Run(":8081")
}

// // collectHandler godoc
// // @Summary Collect SkyWalking payload
// // @Description Receives SkyWalking agent payload and returns it
// // @Accept json
// // @Produce json
// // @Param payload body skywalking.TraceSegment true "Trace Segment"
// // @Success 200 {object} skywalking.TraceSegment
// // @Router /collect [post]
// func collectHandler(c *gin.Context) {
// 	var payload skywalking.TraceSegment
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// Log the payload (or process it)
// 	c.JSON(http.StatusOK, payload)
// }

// // collectOtelHandler godoc
// // @Summary Collect OpenTelemetry payload
// // @Description Receives OTLP trace payload and returns it
// // @Accept json
// // @Produce json
// // @Param payload body otel.OTelPayload true "OTel Trace Payload"
// // @Success 200 {object} otel.OTelPayload
// // @Router /collect-otel [post]
// func collectOtelHandler(c *gin.Context) {
// 	var payload otel.OTelPayload
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if err := SendToOtelBackend(payload); err != nil {
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(200, payload)
// }

// reportPropertiesHandler godoc
// @Summary Handle SkyWalking agent property registration
// @Description Receives agent service/instance properties
// @Accept json
// @Produce json
// @Param payload body InstanceProperties true "Instance Properties"
// @Success 200 {object} map[string]string
// @Router /v3/management/reportProperties [post]
func reportPropertiesHandler(c *gin.Context) {
	fmt.Printf("Received  call")

	// here you could store them, or just acknowledge
	c.JSON(200, gin.H{"status": "received"})
}

// // convertSkywalkingToOtelHandler godoc
// // @Summary Convert SkyWalking trace to OTel format
// // @Description Receives a SkyWalking payload and returns an OTLP-style payload
// // @Accept json
// // @Produce json
// // @Param payload body skywalking.TraceSegment true "SkyWalking Trace Segment"
// // @Success 200 {object} otel.OTelPayload
// // @Router /convert-skywalking-to-otel [post]
// func convertSkywalkingToOtelHandler(c *gin.Context) {
// 	var sw skywalking.TraceSegment
// 	if err := c.ShouldBindJSON(&sw); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

//		// Perform mapping
//		otelPayload := converter.SkywalkingToOtel(sw)
//		c.JSON(200, otelPayload)
//	}
//

// collectAndSendToOtelHandler godoc
// @Summary Collect SkyWalking payload and forward to OTEL backend
// @Description Receives SkyWalking agent payload, converts to OTel, forwards to OTEL backend
// @Accept json
// @Produce json
// @Param payload body skywalking.TraceSegment true "SkyWalking Trace Segment"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v3/segments [post]
func collectAndSendToOtelHandler(c *gin.Context) {
	fmt.Println("execute collectAndSendToOtelHandler")

	// log raw body for debugging
	body, _ := c.GetRawData()
	fmt.Println("RAW BODY:")
	fmt.Println(string(body))

	// reassign the request body so ShouldBindJSON still works
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var payload []skywalking.TraceSegment
	if err := c.ShouldBindJSON(&payload); err != nil {
		fmt.Println("Bind error:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Parsed %d segments\n", len(payload))

	for _, segment := range payload {
		fmt.Printf("Processing segment: traceID=%s\n", segment.TraceID)

		otelPayload := converter.SkywalkingToOtel(segment)

		if err := SendToOtelBackend(otelPayload); err != nil {
			fmt.Printf("Failed to send to OTEL backend: %v\n", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	fmt.Println("All segments processed OK.")
	c.JSON(200, gin.H{"status": "ok"})
}

func SendToOtelBackend(otelPayload otel.OTelPayload) error {
	payloadBytes, err := json.Marshal(otelPayload)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://labs.codexray.io:8041/v1/traces", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("otel backend returned status: %s", resp.Status)
	}

	return nil
}

type InstanceProperties struct {
	ServiceId         string         `json:"serviceId"`
	ServiceInstanceId string         `json:"serviceInstanceId"`
	Properties        map[string]any `json:"properties"`
}

// keepAliveHandler godoc
// @Summary Dummy keepAlive endpoint
// @Description Dummy endpoint to satisfy SkyWalking agent
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /v3/management/keepAlive [post]
func keepAliveHandler(c *gin.Context) {
	fmt.Println("Received keepAlive ping")
	c.JSON(200, gin.H{"status": "alive"})
}

// clrMetricReportsHandler godoc
// @Summary Dummy CLR metrics endpoint
// @Description Dummy endpoint to satisfy SkyWalking .NET agent CLR metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /v3/clrMetricReports [post]
func clrMetricReportsHandler(c *gin.Context) {
	fmt.Println("Received CLR metrics report")
	c.JSON(200, gin.H{"status": "ok"})
}
