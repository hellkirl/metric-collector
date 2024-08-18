package server

import (
	"devops_analytics/internal/storage"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World!"))
}

func MetricsHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		metricsType := r.URL.Query().Get("type")

		tmplPath := filepath.Join("static", "metrics.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			http.Error(w, "Unable to load template", http.StatusInternalServerError)
			return
		}

		metrics := ms.GetMetrics()
		data := make(map[string]any)

		switch metricsType {
		case "gauge":
			for metricName, metricValue := range metrics.Gauges {
				data[metricName] = metricValue
			}
		case "counter":
			for metricName, metricValue := range metrics.Counters {
				data[metricName] = float64(metricValue)
			}
		}

		if err = tmpl.Execute(w, data); err != nil {
			http.Error(w, "Unable to render template", http.StatusInternalServerError)
		}
	}
}

func UpdateMetricHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		var (
			metricType, metricName string
			metricValue            float64
			err                    error
		)

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			http.Error(w, "Invalid path. The path should follow this structure: ~/metricType/metricName/metricValue", http.StatusBadRequest)
			return
		}

		metricType, metricName = parts[0], parts[1]
		metricValue, err = strconv.ParseFloat(parts[2], 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		switch metricType {
		case string(storage.Gauge):
			ms.UpdateGauge(metricName, metricValue)
		case string(storage.Counter):
			ms.UpdateCounter(metricName, int64(metricValue))
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
	}
}
