package main

import (
	"context"
	"sync"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/storage"
)

func runPipeline(ctx context.Context, cancel context.CancelFunc, cfg *Config) {
	// 1. Initial Load of Unprocessed Items
	unprocessedItems, err := storage.GetUnprocessed()
	if err != nil {
		logger.Warn("Failed to load initial unprocessed items: %v", err)
	} else if len(unprocessedItems) > 0 {
		logger.Info("Loaded %d unprocessed items from local storage.", len(unprocessedItems))
	}

	// Channels
	scrapeQueue := make(chan models.Regulation, 60)
	processQueue := make(chan models.Regulation, len(unprocessedItems)+len(scrapeQueue))
	resultQueue := make(chan models.Regulation, len(unprocessedItems)+len(scrapeQueue))

	logger.Section("Execution")
	var wg sync.WaitGroup

	// 2. Start Processor concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		if cfg.ProcessFlag && ctx.Err() == nil {
			runProcessor(ctx, cancel, cfg.ProcessFlag, cfg.APIKey, cfg.ModelFlag, cfg.PromptTemplate, processQueue, resultQueue)
		} else {
			close(processQueue)
			close(resultQueue)
		}
	}()

	// 3. Feeder Goroutine: DB Loading + Scraping
	wg.Add(1)
	go func() {
		defer wg.Done()
		startFeeder(ctx, cfg.ScrapeFlag, unprocessedItems, scrapeQueue, processQueue)
	}()

	// Safety Drain
	startSafetyDrain(ctx, processQueue)

	// 4. Result Collector
	allRegulations := collectResults(cfg, resultQueue)

	wg.Wait()

	// 5. Final Report
	if cfg.FormatFlag != "markdown" {
		logger.Section("Report Generation")
		runReporter(allRegulations, cfg.FormatFlag, cfg.OutputFlag)
	}
}
