package exporter

import (
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
)

// MarkdownStdoutExporter renders regulations in Markdown to the terminal
type MarkdownStdoutExporter struct{}

func (e *MarkdownStdoutExporter) Export(data []models.Regulation, _ string) error {
	processor.RenderMarkdown(data)
	return nil
}
