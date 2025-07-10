package skywalking

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
	IsError       bool        `json:"isError"`
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
