package utils

import (
	"os"
	"testing"
)

func TestStartMetrics(t *testing.T) {
	os.Setenv("METRICS_PORT", "3306")
	StartMetrics()
}
