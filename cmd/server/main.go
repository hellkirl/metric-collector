package main

import (
	serverHandlers "devops_analytics/internal/handlers/server"
	"devops_analytics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	err := run(setupHandler())
	if err != nil {
		panic(err)
	}
}

func run(handler *chi.Mux) error {
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		return err
	}
	return nil
}

func setupHandler() *chi.Mux {
	metricsStorage := storage.NewMemStorageHandler()

	r := chi.NewRouter()

	r.Get("/", serverHandlers.HomePage(metricsStorage))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", serverHandlers.UpdateMetricHandler(metricsStorage))
	r.Get("/value/{metricType}/{metricName}", serverHandlers.MetricsHandler(metricsStorage))

	return r
}
