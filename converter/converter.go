package converter

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"skywalking_transformer/otel"
	"skywalking_transformer/skywalking"
	"strings"
)

func randomTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Failed to generate random trace ID: %v", err)
		// Fallback to a default value
		return ""
	}
	return hex.EncodeToString(b)
}

func randomSpanID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Failed to generate random span ID: %v", err)
		// Fallback to a default value
		return ""
	}
	return hex.EncodeToString(b)
}

// struct to parse the service field
type serviceParsed struct {
	Name   string `json:"name"`
	TeamID string `json:"teamID"`
	Type   string `json:"type"`
}

func SkywalkingToOtel(sw *skywalking.TraceSegment) otel.OTelPayload {
	traceID := randomTraceID()
	spanIDMap := make(map[int]string)
	var otelSpans []otel.OTelSpan

	// parse the sw.Service string
	corrected := strings.ReplaceAll(sw.Service, "'", "\"")
	var parsed serviceParsed
	if err := json.Unmarshal([]byte(corrected), &parsed); err != nil {
		log.Printf("Failed to parse service field, fallback to raw: %v", err)
		parsed.Name = sw.Service
	}

	for i := range sw.Spans {
		swSpan := &sw.Spans[i]
		hexSpanID := randomSpanID()
		spanIDMap[swSpan.SpanID] = hexSpanID

		parentHexID := ""
		if swSpan.ParentSpanID >= 0 {
			parentHexID = spanIDMap[swSpan.ParentSpanID]
		}

		var attributes []otel.Attribute

		// tags as attributes
		for _, tag := range swSpan.Tags {
			if tag.Key == "db.type" {
				tag.Key = "db.system"
			}
			attributes = append(attributes, otel.Attribute{
				Key: tag.Key,
				Value: otel.AttributeVal{
					StringValue: tag.Value,
				},
			})
		}

		// enrich attributes with peer, component, layer
		if swSpan.Peer != "" || swSpan.ComponentId != 0 || swSpan.SpanLayer != "" {
			if swSpan.Peer != "" {
				attributes = append(attributes, otel.Attribute{
					Key: "peer",
					Value: otel.AttributeVal{
						StringValue: swSpan.Peer,
					},
				})
			}
			if swSpan.ComponentId != 0 {
				attributes = append(attributes, otel.Attribute{
					Key: "component.id",
					Value: otel.AttributeVal{
						IntValue: int64(swSpan.ComponentId),
					},
				})
			}
			if swSpan.SpanLayer != "" {
				attributes = append(attributes, otel.Attribute{
					Key: "layer",
					Value: otel.AttributeVal{
						StringValue: swSpan.SpanLayer,
					},
				})
			}
		}

		// Add span type and error status in one append
		attributes = append(attributes, otel.Attribute{
			Key: "span.type",
			Value: otel.AttributeVal{
				StringValue: swSpan.SpanType,
			},
		}, otel.Attribute{
			Key: "span.isError",
			Value: otel.AttributeVal{
				BoolValue: bool(swSpan.IsError),
			},
		})

		// logs as OTel events
		var events []otel.Event
		for k := range swSpan.Logs {
			swLog := &swSpan.Logs[k]
			var eventAttributes []otel.Attribute
			for l := range swLog.Data {
				tag := &swLog.Data[l]
				eventAttributes = append(eventAttributes, otel.Attribute{
					Key: tag.Key,
					Value: otel.AttributeVal{
						StringValue: tag.Value,
					},
				})
			}
			events = append(events, otel.Event{
				Name:         "log",
				TimeUnixNano: formatNano(swLog.Time),
				Attributes:   eventAttributes,
			})
		}

		otelSpan := otel.OTelSpan{
			TraceID:           traceID,
			SpanID:            hexSpanID,
			ParentSpanID:      parentHexID,
			Name:              swSpan.OperationName,
			Kind:              otel.MapSpanTypeToKind(swSpan.SpanType),
			StartTimeUnixNano: formatNano(swSpan.StartTime),
			EndTimeUnixNano:   formatNano(swSpan.EndTime),
			Attributes:        attributes,
			Events:            events,
		}

		otelSpans = append(otelSpans, otelSpan)
	}

	resourceSpans := otel.ResourceSpan{
		Resource: otel.Resource{
			Attributes: []otel.Attribute{
				{
					Key: "service.name",
					Value: otel.AttributeVal{
						StringValue: parsed.Name,
					},
				},
				{
					Key: "service.instance.id",
					Value: otel.AttributeVal{
						StringValue: sw.ServiceInstance,
					},
				},
			},
		},
		ScopeSpans: []otel.ScopeSpans{
			{
				Spans: otelSpans,
			},
		},
	}

	return otel.OTelPayload{
		ResourceSpans: []otel.ResourceSpan{resourceSpans},
	}
}

func formatNano(ms int64) string {
	return fmt.Sprintf("%d", ms*1_000_000)
}
