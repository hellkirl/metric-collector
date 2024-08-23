package server

import (
	"devops_analytics/internal/storage"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFiles("static/index.html", "static/index.css")
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

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		switch metricType {
		case string(storage.Gauge):
			parsedFloatMetricValue, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge metric value", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metricName, parsedFloatMetricValue)
		case string(storage.Counter):
			parsedInt64MetricValue, err := strconv.Atoi(metricValue)
			if err != nil {
				http.Error(w, "Invalid counter metric value", http.StatusBadRequest)
				return
			}
			ms.UpdateCounter(metricName, int64(parsedInt64MetricValue))
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}
