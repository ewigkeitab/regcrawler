package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/storage"
)

func main() {
	cfg := parseConfig()

	logger.Title("RegCrawler - Regulatory Scraper & Summarizer")
	logger.Section("Initialization")

	// Init DB
	if err := storage.InitDB(); err != nil {
		logger.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}

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

	runPipeline(ctx, cancel, cfg)
}
