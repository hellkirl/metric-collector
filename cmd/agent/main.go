package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var (
	endpoint       string
	reportInterval int64
	pollInterval   int64
)

type Metrics struct {
	runtime.MemStats
	RandomValue uint64
	PollCount   uint64
}

func init() {
	flag.StringVar(&endpoint, "a", "localhost:8080", "Endpoint to send metrics")
	flag.Int64Var(&reportInterval, "r", 10, "Frequency of reporting metrics in seconds")
	flag.Int64Var(&pollInterval, "p", 2, "Frequency of polling metrics in seconds")

	flag.Parse()
}

func (m *Metrics) FillFromMemStats(memStats *runtime.MemStats) {
	*m = Metrics{MemStats: *memStats}
	m.RandomValue = rand.Uint64()
}

func (m *Metrics) ToMap(isGauge bool) map[string]any {
	res := map[string]any{
		"Alloc":         m.Alloc,
		"BuckHashSys":   m.BuckHashSys,
		"Frees":         m.Frees,
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         m.GCSys,
		"HeapAlloc":     m.HeapAlloc,
		"HeapIdle":      m.HeapIdle,
		"HeapInuse":     m.HeapInuse,
		"HeapObjects":   m.HeapObjects,
		"HeapReleased":  m.HeapReleased,
		"HeapSys":       m.HeapSys,
		"LastGC":        m.LastGC,
		"Lookups":       m.Lookups,
		"MCacheInuse":   m.MCacheInuse,
		"MCacheSys":     m.MCacheSys,
		"MSpanInuse":    m.MSpanInuse,
		"MSpanSys":      m.MSpanSys,
		"Mallocs":       m.Mallocs,
		"NextGC":        m.NextGC,
		"NumForcedGC":   m.NumForcedGC,
		"NumGC":         m.NumGC,
		"OtherSys":      m.OtherSys,
		"PauseTotalNs":  m.PauseTotalNs,
		"StackInuse":    m.StackInuse,
		"StackSys":      m.StackSys,
		"Sys":           m.Sys,
		"TotalAlloc":    m.TotalAlloc,
	}

	if isGauge {
		res["RandomValue"] = m.RandomValue
	} else {
		m.PollCount++
		res["PollCount"] = m.PollCount
	}

	return res
}

func sendMetrics(metricType string, metrics map[string]any) {
	for metricName, metricValue := range metrics {
		_, err := http.Post(fmt.Sprintf("http://%s/update/%s/%s/%v", endpoint, metricType, metricName, metricValue), "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func startAgent() {
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			var memStats runtime.MemStats
			var metrics Metrics
			runtime.ReadMemStats(&memStats)
			metrics.FillFromMemStats(&memStats)
			sendMetrics("counter", metrics.ToMap(false))
		case <-reportTicker.C:
			var memStats runtime.MemStats
			var metrics Metrics
			runtime.ReadMemStats(&memStats)
			metrics.FillFromMemStats(&memStats)
			sendMetrics("gauge", metrics.ToMap(true))
		}
	}
}

func main() {
	go startAgent()
	select {}
}
