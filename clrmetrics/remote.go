package clrmetrics

import (
	"bytes"
	"net/http"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/klauspost/compress/snappy"
	"github.com/prometheus/prometheus/prompb"
)

var clrCollectorURL = os.Getenv("CLR_REMOTEWRITE_URL")

func remoteWriteCLR(series []prompb.TimeSeries) error {
	if clrCollectorURL == "" {
		clrCollectorURL= "http://demo.codexray.io/v1/metrics" 
	}

	req := &prompb.WriteRequest{Timeseries: series}
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	compressed := snappy.Encode(nil, data)

	httpReq, err := http.NewRequest("POST", clrCollectorURL, bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return err
	}
	return nil
}
