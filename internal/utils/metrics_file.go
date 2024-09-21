package utils

import (
	"devops_analytics/internal/logger"
	"devops_analytics/internal/storage"
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

func RestoreMetrics(restore bool, fileStoragePath string, ms *storage.MemStorage) {
	if restore {
		file, err := os.OpenFile(fileStoragePath, os.O_RDONLY, 0644)
		if err != nil {
			logger.Log.Error("Error opening file: ", err)
			return
		}
		defer file.Close()

		var data []byte
		_, err = file.Read(data)
		ms.FromJson(data)
	}
}
