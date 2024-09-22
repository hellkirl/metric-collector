package utils

import (
	"devops_analytics/internal/logger"
	"io"
	"os"
)

func SaveMetrics(fileStoragePath string, data []byte) {
	file, err := os.OpenFile(fileStoragePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Error("Error opening file: ", err)
		return
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		logger.Log.Error("Error writing to file: ", err)
		return
	}
}

func RestoreMetrics(fileStoragePath string) []byte {
	file, err := os.OpenFile(fileStoragePath, os.O_RDONLY, 0644)
	if err != nil {
		logger.Log.Error("Error opening file: ", err)
		return nil
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		logger.Log.Error("Error reading from file: ", err)
		return nil
	}
	return data
}
