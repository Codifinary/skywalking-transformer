package otel

// OTelPayload models the OTLP JSON structure for traces.
type OTelPayload struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

type ResourceSpan struct {
	Resource   Resource     `json:"resource"`
	ScopeSpans []ScopeSpans `json:"scopeSpans"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type ScopeSpans struct {
	Spans []OTelSpan `json:"spans"`
}

type OTelSpan struct {
	TraceID           string      `json:"traceId"`
	SpanID            string      `json:"spanId"`
	ParentSpanID      string      `json:"parentSpanId"`
	Name              string      `json:"name"`
	Kind              string      `json:"kind"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	EndTimeUnixNano   string      `json:"endTimeUnixNano"`
	Attributes        []Attribute `json:"attributes"`
	Events            []Event     `json:"events,omitempty"`
}

type Attribute struct {
	Key   string       `json:"key"`
	Value AttributeVal `json:"value"`
}

type AttributeVal struct {
	StringValue string `json:"stringValue,omitempty"`
	IntValue    int64  `json:"intValue,omitempty"`
	BoolValue   bool   `json:"boolValue,omitempty"`
}

type Event struct {
	Name         string      `json:"name"`
	TimeUnixNano string      `json:"timeUnixNano"`
	Attributes   []Attribute `json:"attributes"`
}

func MapSpanTypeToKind(spanType string) string {
	switch spanType {
	case "0":
		return "SPAN_KIND_SERVER"
	case "1":
		return "SPAN_KIND_CLIENT"
	case "2":
		return "SPAN_KIND_INTERNAL"
	default:
		return "SPAN_KIND_INTERNAL"
	}
}
