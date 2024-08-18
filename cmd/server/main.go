package main

import (
	serverHandlers "devops_analytics/internal/handlers/server"
	"devops_analytics/internal/storage"
	"net/http"
)

func main() {
	err := run(setupHandler())
	if err != nil {
		panic(err)
	}
}

func run(handler *http.ServeMux) error {
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		return err
	}
	return nil
}

func setupHandler() *http.ServeMux {
	metricsStorage := storage.NewMemStorageHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/", serverHandlers.HomePage)
	mux.HandleFunc("/update/", serverHandlers.UpdateMetricHandler(metricsStorage))
	mux.HandleFunc("/metrics", serverHandlers.MetricsHandler(metricsStorage))
	return mux
}
