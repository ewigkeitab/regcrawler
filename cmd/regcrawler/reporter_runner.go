package main

import (
	"regcrawler/pkg/exporter"
	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
)

func runReporter(allRegulations []models.Regulation, formatFlag, outputFlag string) {
	if len(allRegulations) > 0 {
		logger.Success("Collected %d regulations.", len(allRegulations))

		var exp exporter.Exporter
		switch formatFlag {
		case "json":
			exp = &exporter.JSONExporter{}
		case "markdown":
			exp = &exporter.MarkdownFileExporter{}
		case "mdstdout":
			exp = &exporter.MarkdownStdoutExporter{}
		default:
			logger.Warn("Unknown format: %s. Defaulting to markdown.", formatFlag)
			exp = &exporter.MarkdownFileExporter{}
		}

		if err := exp.Export(allRegulations, outputFlag); err != nil {
			logger.Error("Failed to export: %v", err)
		}
	} else {
		logger.Warn("No regulations collected.")
	}
}
