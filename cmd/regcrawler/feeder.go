package main

import (
	"context"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/storage"
)

func startFeeder(ctx context.Context, scrapeFlag bool, unprocessedItems []models.Regulation, scrapeQueue, processQueue chan models.Regulation) {
	defer close(processQueue)

	// A. Inject previously unprocessed items
	for _, item := range unprocessedItems {
		select {
		case <-ctx.Done():
			return
		case processQueue <- item:
		}
	}

	// B. Run Scraper
	if scrapeFlag {
		runScraper(ctx, scrapeFlag, scrapeQueue)
		// C. Pipe Scraper -> DB + Processor
		for reg := range scrapeQueue {
			// Save immediately to DB (INSERT OR IGNORE)
			if err := storage.SaveUnprocessed(reg); err != nil {
				logger.Error("Failed to save to database %s: %v", reg.Link, err)
			}
			// Feed to processor
			select {
			case <-ctx.Done():
				return
			case processQueue <- reg:
			}
		}
	} else {
		close(scrapeQueue)
	}
}
