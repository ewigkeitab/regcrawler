package exporter

import (
	"encoding/json"
	"os"
	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
)

// JSONExporter exports regulations to a JSON file
type JSONExporter struct{}

func (e *JSONExporter) Export(data []models.Regulation, filename string) error {
	if filename == "" {
		filename = "processed_regulations.json"
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return err
	}

	logger.Success("Data saved to %s", filename)
	return nil
}
