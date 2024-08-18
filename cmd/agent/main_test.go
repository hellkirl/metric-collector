package main

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"runtime"
	"testing"
)

func TestMetrics_FillFromMemStats_Gauge(t *testing.T) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var metrics Metrics
	randVal := rand.Uint64()

	metrics.RandomValue = randVal
	metrics.FillFromMemStats(&memStats)

	assert.NotEmpty(t, metrics.RandomValue, "Gauge Metrics RandomValue shouldn't be empty")
	assert.Zero(t, metrics.PollCount, "Gauge Metrics PollCount should be zero")

	memStatsValue := reflect.ValueOf(memStats)
	memStatsType := reflect.TypeOf(memStats)

	metricsMap := metrics.ToMap(true)

	for i := 0; i < memStatsValue.NumField(); i++ {
		fieldName := memStatsType.Field(i).Name
		expectedValue := memStatsValue.Field(i).Interface()

		if actualValue, ok := metricsMap[fieldName]; ok {
			assert.Equal(t, expectedValue, actualValue, "Field %s mismatch", fieldName)
		}
	}
}
