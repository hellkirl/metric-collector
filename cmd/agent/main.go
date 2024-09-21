package main

import (
	"devops_analytics/internal/logger"
	"devops_analytics/internal/models"
	"devops_analytics/internal/utils"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
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
}

var (
	cfg           Config
	metrics       MemStats
	rwMu          sync.RWMutex
	memStats      runtime.MemStats
	pollCount     atomic.Int64
	lastPollCount int64 = 0
)

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
	rwMu.Lock()
	m.MemStats = *memStats
	m.RandomValue = rand.Uint64()
	rwMu.Unlock()
	pollCount.Add(1)
}

func (m *MemStats) ToGaugeMap() map[string]float64 {
	rwMu.Lock()
	defer rwMu.Unlock()
	return map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
		"RandomValue":   float64(m.RandomValue),
	}
}

func (m *MemStats) ToCounterMap() map[string]uint64 {
	currentCount := pollCount.Load()
	delta := currentCount - lastPollCount
	lastPollCount = currentCount

	return map[string]uint64{
		"PollCount": uint64(delta),
	}
}

func sendMetrics(gaugeMetrics map[string]float64, counterMetrics map[string]uint64) {
	for metricName, metricValue := range gaugeMetrics {
		go func(metricName string, metricValue float64) {
			var body models.AgentMetrics

			body.MType = "gauge"
			body.ID = metricName
			body.Value = &metricValue

			compressedBody := utils.CompressBody(body)
			if compressedBody == nil {
				if logger.Log != nil {
					logger.Log.Error("Failed to compress body for metric:", metricName)
				}
				return
			}

			_, err := resty.New().R().
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Content-Type", "application/json").
				SetBody(compressedBody).
				Post(fmt.Sprintf("http://%s/update", cfg.Address))

			if err != nil {
				if logger.Log != nil {
					logger.Log.Error("Couldn't send metrics:", err)
				}
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range counterMetrics {
		go func(metricName string, metricValue uint64) {
			var body models.AgentMetrics

			body.MType = "counter"
			body.ID = metricName
			body.Delta = &metricValue

			compressedBody := utils.CompressBody(body)
			if compressedBody == nil {
				if logger.Log != nil {
					logger.Log.Error("Failed to compress body for metric:", metricName)
				}
				return
			}

			_, err := resty.New().R().
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Content-Type", "application/json").
				SetBody(compressedBody).
				Post(fmt.Sprintf("http://%s/update", cfg.Address))

			if err != nil {
				if logger.Log != nil {
					logger.Log.Error("Couldn't send metrics:", err)
				}
			}
		}(metricName, metricValue)
	}
}

func startAgent() {
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)

	for {
		select {
		case <-pollTicker.C:
			runtime.ReadMemStats(&memStats)
			metrics.FillFromMemStats(&memStats)
		case <-reportTicker.C:
			sendMetrics(metrics.ToGaugeMap(), metrics.ToCounterMap())
		}
	}
}

func main() {
	go startAgent()
	select {}
}
