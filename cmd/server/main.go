package main

import (
	serverHandlers "devops_analytics/internal/handlers/server"
	customMiddleware "devops_analytics/internal/middleware"
	"devops_analytics/internal/storage"
	"devops_analytics/internal/utils"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	StorageInterval int64  `env:"STORAGE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

var (
	cfg            Config
	metricsStorage *storage.MemStorage
)

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Println("Couldn't parse env variables. Falling back to flags...")
	}

	if cfg.StorageInterval == 0 {
		flag.Int64Var(&cfg.StorageInterval, "i", 5, "New data storage interval")
	}
	if cfg.FileStoragePath == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		defaultPath := filepath.Join(workingDir, "tmp/metrics-db.json")
		flag.StringVar(&cfg.FileStoragePath, "f", defaultPath, "File to store new data in")
	}
	if cfg.Restore == false {
		flag.BoolVar(&cfg.Restore, "r", false, "Restore data when starting the server")
	}

	flag.Parse()

	metricsStorage = storage.NewMemStorageHandler()
}

func main() {
	if cfg.Restore {
		previousMetrics := utils.RestoreMetrics(cfg.FileStoragePath)
		metricsStorage.FromJson(previousMetrics)
	}

	go storeMetrics()

	err := run(setupHandler())
	if err != nil {
		panic(err)
	}
}

func storeMetrics() {
	storageTicker := time.NewTicker(time.Duration(cfg.StorageInterval) * time.Second)
	for {
		select {
		case <-storageTicker.C:
			utils.SaveMetrics(cfg.FileStoragePath, metricsStorage.ToJson())
		}
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
	r := chi.NewRouter()
	r.Use(customMiddleware.LoggerMiddleware, customMiddleware.GzipRequest, middleware.SetHeader("Accept-Encoding", "gzip"), customMiddleware.GzipResponse)

	r.Get("/", serverHandlers.HomePage(metricsStorage))
	r.Post("/update", serverHandlers.UpdateMetricHandler(metricsStorage))
	r.Get("/value/{metricType}/{metricName}", serverHandlers.MetricsHandler(metricsStorage))

	return r
}
