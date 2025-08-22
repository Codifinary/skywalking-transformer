package skywalking

import (
	"encoding/json"
	"fmt"
)

type BoolInt bool

func (b *BoolInt) UnmarshalJSON(data []byte) error {
	// Try boolean
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		*b = BoolInt(boolVal)
		return nil
	}

	// Try int (0 or 1)
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*b = BoolInt(intVal != 0)
		return nil
	}

	return fmt.Errorf("invalid value for BoolInt: %s", string(data))
}

type TraceSegment struct {
	TraceID         string `json:"traceId"`
	Service         string `json:"service"`
	ServiceInstance string `json:"serviceInstance"`
	TraceSegmentId  string `json:"traceSegmentId"`
	IsSizeLimited   bool   `json:"isSizeLimited"`
	Spans           []Span `json:"spans"`
}

type Span struct {
	SpanID        int         `json:"spanId"`
	ParentSpanID  int         `json:"parentSpanId"`
	OperationName string      `json:"operationName"`
	SpanType      string      `json:"spanType"`
	IsError       BoolInt     `json:"isError"`
	StartTime     int64       `json:"startTime"`
	EndTime       int64       `json:"endTime"`
	Peer          string      `json:"peer"`
	ComponentId   int         `json:"componentId"`
	SpanLayer     string      `json:"spanLayer"`
	SkipAnalysis  bool        `json:"skipAnlysis"`
	Tags          []Tag       `json:"tags"`
	Logs          []Log       `json:"logs"`
	References    []Reference `json:"references"`
	MethodName    *string     `json:"methodName"` // nullable
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Log struct {
	Time int64 `json:"time"`
	Data []Tag `json:"data"`
}

type Reference struct {
	TraceId string `json:"traceId"`
	Headers string `json:"headers"`
}
