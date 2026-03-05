package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
	"regcrawler/pkg/storage"
)

func collectResults(cfg *Config, resultQueue <-chan models.Regulation) []models.Regulation {
	var allRegulations []models.Regulation
	outputFile := cfg.OutputFlag
	if outputFile == "" && cfg.FormatFlag == "markdown" {
		outputFile = "regulatory_report.md"
	}

	var f *os.File
	if cfg.FormatFlag == "markdown" && outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error("Failed to open output file: %v", err)
		} else {
			defer f.Close()
			// If file is new or empty, write header
			info, _ := f.Stat()
			if info.Size() == 0 {
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				header := fmt.Sprintf("# 最新法規動態彙整\n整理時間: %s\n\n", timestamp)
				if _, err := f.WriteString(header); err != nil {
					logger.Error("Failed to write header: %v", err)
				}
			}
		}
	}

	for reg := range resultQueue {
		allRegulations = append(allRegulations, reg)

		// Immediate writing for markdown
		if cfg.FormatFlag == "markdown" && f != nil {
			md := processor.GenerateItemMarkdown(reg)
			if _, err := f.WriteString(md); err != nil {
				logger.Error("Failed to write item to markdown: %v", err)
			}
		}

		// Immediate DB marking if successfully processed
		if reg.Keypoints != "" && !strings.HasPrefix(reg.Keypoints, "[warning]") && !strings.HasPrefix(reg.Keypoints, "Error") {
			if err := storage.DeleteProcessed(reg.Link); err != nil {
				logger.Error("Failed to remove processed item from database: %v", err)
			}
			if err := storage.MarkProcessed(reg.Link); err != nil {
				logger.Error("Failed to mark item as fully processed in database: %v", err)
			}
			logger.Info("Marked %s as processed in database.", reg.Title)
		}
	}

	if cfg.FormatFlag == "markdown" {
		logger.Success("Incremental Markdown report updated in %s", outputFile)
	}

	return allRegulations
}
