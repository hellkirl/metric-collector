package server

import (
	"devops_analytics/internal/models"
	"devops_analytics/internal/storage"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"log"
	"net/http"
)

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFiles("static/index.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
}

func HomePage(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		msData := ms.GetMetrics()
		data := struct {
			Gauges   map[string]float64
			Counters map[string]int64
		}{
			Gauges:   msData.Gauges,
			Counters: msData.Counters,
		}

		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Failed to execute template: %v", err)
			return
		}
	}
}

func MetricsHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		metrics := ms.GetMetrics()

		switch metricType {
		case string(storage.Gauge):
			if val, ok := metrics.Gauges[metricName]; ok {
				w.Write([]byte(fmt.Sprintf("%f", val)))
			} else {
				http.Error(w, "Unknown metric name", http.StatusNotFound)
			}
		case string(storage.Counter):
			if val, ok := metrics.Counters[metricName]; ok {
				w.Write([]byte(fmt.Sprintf("%d", val)))
			} else {
				http.Error(w, "Unknown metric name", http.StatusNotFound)
			}
		default:
			http.Error(w, "Unknown metric type", http.StatusNotFound)
		}
	}
}

func UpdateMetricHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			metric models.Metrics
			body   []byte
			err    error
		)

		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Couldn't read request body", http.StatusBadRequest)
			log.Printf("Error reading request body: %v", err)
			return
		}
		defer r.Body.Close()

		if err = json.Unmarshal(body, &metric); err != nil {
			fmt.Println(string(body))
			http.Error(w, "Invalid JSON format for metrics", http.StatusBadRequest)
			log.Printf("JSON unmarshal error: %v", err)
			return
		}

		switch metric.MType {
		case string(storage.Gauge):
			if metric.Value == nil {
				http.Error(w, "Value is required for gauge", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metric.ID, *metric.Value)
		case string(storage.Counter):
			if metric.Delta == nil {
				http.Error(w, "Delta is required for counter", http.StatusBadRequest)
				return
			}
			ms.UpdateCounter(metric.ID, *metric.Delta)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
