package storage

import (
	"sync"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MemStorage struct {
	mu       sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

type ReadOnlyStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorageHandler() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(name string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = value
}

func (ms *MemStorage) UpdateCounter(name string, value int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += value
}

func (ms *MemStorage) GetMetrics() ReadOnlyStorage {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	resp := ReadOnlyStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	for k, v := range ms.gauges {
		resp.Gauges[k] = v
	}
	for k, v := range ms.counters {
		resp.Counters[k] = v
	}

	return resp
}
