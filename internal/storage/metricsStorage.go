package storage

import (
	"devops_analytics/internal/logger"
	"encoding/json"
	"sync"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MemStorage struct {
	rwm      sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

type ReadOnlyStorage struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewMemStorageHandler() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(name string, value float64) {
	ms.rwm.Lock()
	ms.gauges[name] = value
	ms.rwm.Unlock()
}

func (ms *MemStorage) UpdateCounter(name string, value int64) {
	ms.rwm.Lock()
	ms.counters[name] += value
	ms.rwm.Unlock()
}

func (ms *MemStorage) GetMetrics() ReadOnlyStorage {
	ms.rwm.RLock()
	defer ms.rwm.RUnlock()

	return ReadOnlyStorage{
		Gauges:   ms.gauges,
		Counters: ms.counters,
	}
}

func (ms *MemStorage) ToJson() []byte {
	metrics := ms.GetMetrics()
	body, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		logger.Log.Error("Error marshaling metrics to json: ", err)
	}
	return body
}

func (ms *MemStorage) FromJson(data []byte) {
	ms.rwm.Lock()
	defer ms.rwm.Unlock()

	var metrics ReadOnlyStorage
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		logger.Log.Error("Error unmarshaling metrics from json: ", err)
		return
	}

	ms.counters = metrics.Counters
	ms.gauges = metrics.Gauges
}
