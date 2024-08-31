package main

import (
	serverHandlers "devops_analytics/internal/handlers/server"
	customMiddleware "devops_analytics/internal/middleware"
	"devops_analytics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	r.Use(customMiddleware.LoggerMiddleware, customMiddleware.Compress, middleware.SetHeader("Accept-Encoding", "gzip"), customMiddleware.Decompress)

	r.Get("/", serverHandlers.HomePage(metricsStorage))
	r.Post("/update", serverHandlers.UpdateMetricHandler(metricsStorage))
	r.Get("/value/{metricType}/{metricName}", serverHandlers.MetricsHandler(metricsStorage))

	return r
}
