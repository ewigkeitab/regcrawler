package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
	"regcrawler/pkg/storage"
	"strings"
)

func main() {
	scrapeFlag := flag.Bool("scrape", true, "Run the scraper to fetch new regulations")
	processFlag := flag.Bool("process", true, "Run the AI processor to summarize regulations")
	skipAIFlag := flag.Bool("skip-ai", false, "Skip AI processing and just generate report")
	formatFlag := flag.String("format", "markdown", "Output format: markdown or json")
	modelFlag := flag.String("model", "gemini-2.5-flash", "AI Model to use (e.g., gemini-2.0-flash, gemini-2.5-flash)")
	promptFlag := flag.String("prompt", "", "Path to custom prompt text file")
	outputFlag := flag.String("output", "", "Output file name (optional)")

	flag.Parse()

	if *skipAIFlag {
		*processFlag = false
	}

	// Load prompt
	promptTemplate := processor.DefaultPrompt
	if *promptFlag != "" {
		content, err := os.ReadFile(*promptFlag)
		if err != nil {
			logger.Error("Error reading prompt file: %v", err)
			os.Exit(1)
		}
		promptTemplate = string(content)
		logger.Info("Loaded custom prompt from %s", *promptFlag)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if *processFlag && apiKey == "" {
		logger.Error("GEMINI_API_KEY environment variable not set.")
		fmt.Println("Please set it using: export GEMINI_API_KEY='your_key_here'")
		os.Exit(1)
	}

	logger.Title("RegCrawler - Regulatory Scraper & Summarizer")
	logger.Section("Initialization")

	// Init DB
	if err := storage.InitDB(); err != nil {
		logger.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}

	// Load Unprocessed Items
	unprocessedItems, err := storage.GetUnprocessed()
	if err != nil {
		logger.Warn("Failed to load unprocessed items: %v", err)
	} else if len(unprocessedItems) > 0 {
		logger.Info("Loaded %d unprocessed items from local storage to retry.", len(unprocessedItems))
	}

	// Channels
	scrapeQueue := make(chan models.Regulation, 60)                                      // Scraper -> Saver
	processQueue := make(chan models.Regulation, len(unprocessedItems)+len(scrapeQueue)) // Saver -> Processor
	resultQueue := make(chan models.Regulation, len(unprocessedItems)+len(scrapeQueue))  // Processor -> Collector

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Warn("Received interrupt, shutting down...")
		cancel()
	}()

	var wg sync.WaitGroup

	// 1. Start Scraper
	runScraper(ctx, *scrapeFlag, scrapeQueue)

	// 1.5. DB Intermediary (Saver)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// First, inject previously unprocessed items to processQueue
		for _, item := range unprocessedItems {
			processQueue <- item
		}
		// Then, process incoming newly scraped items
		for reg := range scrapeQueue {
			// Save to DB before trying to process
			if err := storage.SaveUnprocessed(reg); err != nil {
				logger.Error("Failed to save to database %s: %v", reg.Link, err)
			}
			processQueue <- reg
		}
		close(processQueue)
	}()

	// 2. Start Processor
	runProcessor(ctx, cancel, *processFlag, apiKey, *modelFlag, promptTemplate, processQueue, resultQueue)

	// 3. Collect Results
	var allRegulations []models.Regulation
	outputFile := *outputFlag
	if outputFile == "" && *formatFlag == "markdown" {
		outputFile = "regulatory_report.md"
	}

	var f *os.File
	if *formatFlag == "markdown" && outputFile != "" {
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
		if *formatFlag == "markdown" && f != nil {
			md := processor.GenerateItemMarkdown(reg)
			if _, err := f.WriteString(md); err != nil {
				logger.Error("Failed to write item to markdown: %v", err)
			}
		}

		// Immediate DB marking if successfully processed
		// Note: We check if it's not a warning/error (matching logic in processor.go)
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

	// Wait for DB Saver to finish all writes cleanly
	wg.Wait()

	// 4. Report
	if *formatFlag != "markdown" {
		logger.Section("Report Generation")
		runReporter(allRegulations, *formatFlag, *outputFlag)
	} else {
		logger.Success("Incremental Markdown report updated in %s", outputFile)
	}
}
