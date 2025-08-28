package clrmetrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/prompb"
)

func CLRHandler(c *gin.Context) {
	var payload Payload
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := payload.Service
	inst := payload.ServiceInstance
	now := time.Now().UnixMilli()

	var ts []prompb.TimeSeries

	for _, m := range payload.Metrics {
		// CPU
		ts = append(ts, buildTS("dotnet_cpu_usage_percent", svc, inst, now, m.CPU.UsagePercent))

		// GC
		ts = append(ts,
			buildTS("dotnet_gc_bytes_in_all_heaps", svc, inst, now, m.GC.BytesInAllHeaps),
			buildTS("dotnet_gc_gen0_collect_count", svc, inst, now, m.GC.Gen0CollectCount),
			buildTS("dotnet_gc_gen1_collect_count", svc, inst, now, m.GC.Gen1CollectCount),
			buildTS("dotnet_gc_gen2_collect_count", svc, inst, now, m.GC.Gen2CollectCount),
			buildTS("dotnet_gc_heap_memory", svc, inst, now, m.GC.HeapMemory),
		)

		// Threads
		ts = append(ts,
			buildTS("dotnet_thread_available_worker_threads", svc, inst, now, m.Thread.AvailableWorkerThreads),
			buildTS("dotnet_thread_available_completion_threads", svc, inst, now, m.Thread.AvailableCompletionPortThreads),
			buildTS("dotnet_thread_total_contentions", svc, inst, now, m.Thread.TotalContentions),
		)

		// Exceptions
		ts = append(ts,
			buildTS("dotnet_exceptions_thrown", svc, inst, now, m.Exception.ExThrown),
			buildTS("dotnet_exceptions_per_sec", svc, inst, now, m.Exception.ExThrownPerSec),
		)

		// JIT
		ts = append(ts,
			buildTS("dotnet_jit_methods_jitted", svc, inst, now, m.JIT.MethodsJitted),
			buildTS("dotnet_jit_time_in_jit", svc, inst, now, m.JIT.TimeInJIT),
		)
	}

	// Remote write
	if err := remoteWriteCLR(ts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "clr metrics sent"})
}

func buildTS(name, svc, inst string, ts int64, value float64) prompb.TimeSeries {
	return prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "__name__", Value: name},
			{Name: "service", Value: svc},
			{Name: "service_instance", Value: inst},
		},
		Samples: []prompb.Sample{{Value: value, Timestamp: ts}},
	}
}
