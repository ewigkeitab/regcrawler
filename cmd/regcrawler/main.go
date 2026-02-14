package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
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

	// Channels
	// Scraper -> scrapeQueue -> Processor/Saver
	scrapeQueue := make(chan models.Regulation, 30)
	// Processor -> resultQueue -> Saver
	resultQueue := make(chan models.Regulation, 30)

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

	logger.Title("RegCrawler - Regulatory Scraper & Summarizer")
	logger.Section("Initialization")

	// 1. Start Scraper
	runScraper(*scrapeFlag, scrapeQueue)

	// 2. Start Processor
	runProcessor(ctx, *processFlag, apiKey, *modelFlag, promptTemplate, scrapeQueue, resultQueue)

	// 3. Collect Results (Saver)
	var allRegulations []models.Regulation
	// logger.Info("Waiting for results...")

	for reg := range resultQueue {
		allRegulations = append(allRegulations, reg)
	}

	// 4. Report
	logger.Section("Report Generation")
	runReporter(allRegulations, *formatFlag, *outputFlag)
}
