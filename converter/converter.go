package converter

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sky2otel/otel"
	"sky2otel/skywalking"
	"strings"
)

func randomTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func randomSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// struct to parse the service field
type serviceParsed struct {
	Name   string `json:"name"`
	TeamID string `json:"teamID"`
	Type   string `json:"type"`
}

func SkywalkingToOtel(sw skywalking.TraceSegment) otel.OTelPayload {
	traceID := randomTraceID()
	spanIDMap := make(map[int]string)
	var otelSpans []otel.OTelSpan

	// parse the sw.Service string
	corrected := strings.ReplaceAll(sw.Service, "'", "\"")
	var parsed serviceParsed
	if err := json.Unmarshal([]byte(corrected), &parsed); err != nil {
		fmt.Printf("Failed to parse service field, fallback to raw: %v\n", err)
		parsed.Name = sw.Service
	}

	for _, swSpan := range sw.Spans {
		hexSpanID := randomSpanID()
		spanIDMap[swSpan.SpanID] = hexSpanID

		parentHexID := ""
		if swSpan.ParentSpanID >= 0 {
			parentHexID = spanIDMap[swSpan.ParentSpanID]
		}

		var attributes []otel.Attribute

		// tags as attributes
		for _, tag := range swSpan.Tags {
			attributes = append(attributes, otel.Attribute{
				Key: tag.Key,
				Value: otel.AttributeVal{
					StringValue: tag.Value,
				},
			})
		}

		// enrich attributes with peer, component, layer
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

		attributes = append(attributes, otel.Attribute{
			Key: "span.type",
			Value: otel.AttributeVal{
				StringValue: swSpan.SpanType,
			},
		})

		attributes = append(attributes, otel.Attribute{
			Key: "span.isError",
			Value: otel.AttributeVal{
				BoolValue: swSpan.IsError,
			},
		})

		// logs as OTel events
		var events []otel.Event
		for _, swLog := range swSpan.Logs {
			var eventAttributes []otel.Attribute
			for _, tag := range swLog.Data {
				eventAttributes = append(eventAttributes, otel.Attribute{
					Key: tag.Key,
					Value: otel.AttributeVal{
						StringValue: tag.Value,
					},
				})
			}
			events = append(events, otel.Event{
				Name:         "log",
				TimeUnixNano: fmt.Sprintf("%d", swLog.Time*1_000_000),
				Attributes:   eventAttributes,
			})
		}

		otelSpan := otel.OTelSpan{
			TraceID:           traceID,
			SpanID:            hexSpanID,
			ParentSpanID:      parentHexID,
			Name:              swSpan.OperationName,
			Kind:              otel.MapSpanTypeToKind(swSpan.SpanType),
			StartTimeUnixNano: fmt.Sprintf("%d", swSpan.StartTime*1_000_000),
			EndTimeUnixNano:   fmt.Sprintf("%d", swSpan.EndTime*1_000_000),
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
