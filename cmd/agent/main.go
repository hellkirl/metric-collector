package main

import (
	"devops_analytics/internal/logger"
	"devops_analytics/internal/models"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"runtime"
	"time"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

type MemStats struct {
	runtime.MemStats
	RandomValue uint64
	PollCount   uint64
}

var cfg Config

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Println("Couldn't parse env variables. Falling back to flags...")
	}

	if cfg.Address == "" {
		flag.StringVar(&cfg.Address, "a", "localhost:8080", "Endpoint to send metrics")
	}
	if cfg.ReportInterval == 0 {
		flag.Int64Var(&cfg.ReportInterval, "r", 10, "Frequency of reporting metrics in seconds")
	}
	if cfg.PollInterval == 0 {
		flag.Int64Var(&cfg.PollInterval, "p", 2, "Frequency of polling metrics in seconds")
	}

	flag.Parse()
}

func (m *MemStats) FillFromMemStats(memStats *runtime.MemStats) {
	*m = MemStats{MemStats: *memStats}
	m.RandomValue = rand.Uint64()
}

func (m *MemStats) ToMap(isGauge bool) map[string]any {
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
		var body models.AgentMetrics

		switch metricType {
		case "gauge":
			body = models.AgentMetrics{
				ID:    "gauge",
				MType: metricName,
				Value: metricValue,
			}
		case "counter":
			body = models.AgentMetrics{
				ID:    "counter",
				MType: metricName,
				Delta: metricValue,
			}
		}

		_, err := resty.New().R().SetHeader("Content-Encoding", "gzip").SetBody(body).Post(fmt.Sprintf("http://%s/update", cfg.Address))
		if err != nil {
			logger.Log.Error("Couldn't send metrics")
		}
	}
}

func startAgent() {
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			var memStats runtime.MemStats
			var metrics MemStats
			runtime.ReadMemStats(&memStats)
			metrics.FillFromMemStats(&memStats)
			sendMetrics("counter", metrics.ToMap(false))
		case <-reportTicker.C:
			var memStats runtime.MemStats
			var metrics MemStats
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
