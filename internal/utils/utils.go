package utils

import (
	"bytes"
	"compress/gzip"
	"devops_analytics/internal/logger"
	"devops_analytics/internal/models"
	"encoding/json"
)

func CompressBody(metric models.AgentMetrics) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(metric); err != nil {
		logger.Log.Error("Couldn't encode metric to JSON:", err)
		gz.Close()
		return nil
	}

	if err := gz.Close(); err != nil {
		logger.Log.Error("Error closing gzip writer:", err)
		return nil
	}

	return b.Bytes()
}
