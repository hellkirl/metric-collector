package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Metrics struct {
	runtime.MemStats
	RandomValue uint64
	PollCount   uint64
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

func sendMetrics(endpoint string, metrics map[string]any) {
	for k, v := range metrics {
		_, err := http.Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", endpoint, k, v), "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func startAgent() {
	const reportInterval = 10 * time.Second
	const pollInterval = 2 * time.Second

	reportTicker := time.NewTicker(reportInterval)
	pollTicker := time.NewTicker(pollInterval)

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
