package exporter

import (
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
)

// MarkdownFileExporter exports regulations to a Markdown file
type MarkdownFileExporter struct{}

func (e *MarkdownFileExporter) Export(data []models.Regulation, filename string) error {
	processor.GenerateMarkdownReport(data, filename)
	return nil
}
