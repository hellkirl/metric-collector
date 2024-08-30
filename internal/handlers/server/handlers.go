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
	"strings"
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
			Gauge   map[string]float64
			Counter map[string]int64
		}{
			Gauge:   msData.Gauges,
			Counter: msData.Counters,
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

		foundMetricNameFlag := false

		switch metricType {
		case string(storage.Gauge):
			for key, val := range metrics.Gauges {
				if strings.ToLower(metricName) == strings.ToLower(key) {
					w.Write([]byte(fmt.Sprintf("%f", val)))
					foundMetricNameFlag = true
					break
				}
			}
			if !foundMetricNameFlag {
				http.Error(w, "Unknown metric name", http.StatusNotFound)
			}
		case string(storage.Counter):
			for key, val := range metrics.Counters {
				if strings.ToLower(metricName) == strings.ToLower(key) {
					w.Write([]byte(fmt.Sprintf("%d", val)))
					foundMetricNameFlag = true
					break
				}
			}
			if !foundMetricNameFlag {
				http.Error(w, "Unknown metric name", http.StatusNotFound)
			}
		default:
			http.Error(w, "Unknown metric type", http.StatusNotFound)
		}
	}
}

func UpdateMetricHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		var metric models.Metrics
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Couldn't handle the request", http.StatusServiceUnavailable)
			return
		}
		defer r.Body.Close()
		if err = json.Unmarshal(body, &metric); err != nil {
			http.Error(w, "Make sure all the metrics are correct", http.StatusBadRequest)
			return
		}

		if (metric.ID == string(storage.Gauge) && metric.Delta != nil) || (metric.ID == string(storage.Counter) && metric.Value != nil) {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		switch metric.ID {
		case string(storage.Gauge):
			if metric.Value == nil {
				http.Error(w, "Value is required for gauge", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metric.MType, *metric.Value)
		case string(storage.Counter):
			if metric.Delta == nil {
				http.Error(w, "Delta is required for counter", http.StatusBadRequest)
				return
			}
			ms.UpdateCounter(metric.MType, *metric.Delta)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
