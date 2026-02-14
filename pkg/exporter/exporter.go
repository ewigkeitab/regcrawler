package exporter

import (
	"regcrawler/pkg/models"
)

// Exporter defines the interface for exporting regulations
type Exporter interface {
	Export(data []models.Regulation, filename string) error
}
