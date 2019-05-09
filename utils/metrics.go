package utils

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

const DefaultMetricPort = "3006"
const DefaultMetricPath = "/metrics"

func StartMetrics() {
	port := os.Getenv("METRICS_PORT")
	if len(port) == 0 {
		port = DefaultMetricPort
	} else {
		p, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}
		if p > 65535 || p < 0 {
			panic("METRICS_PORT must between 0 and 65535 ")
		}
	}

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), MetricsHandler{})
	if err != nil {
		Errorf("metrics service error: %v", err)
	}
}

type MetricsHandler struct {
}

func (MetricsHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.URL.Path != DefaultMetricPath {
		resp.WriteHeader(http.StatusNotFound)
		responseBody(resp, "Not Found")
		return
	}

	responseBody(resp, "Hello")
}

func responseBody(resp http.ResponseWriter, data string) {
	_, err := resp.Write([]byte(data))
	if err != nil {
		Errorf("metrics error: %v", err)
	}
}
