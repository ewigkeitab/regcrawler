package main

import (
	"context"
	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/scraper"
)

func runScraper(ctx context.Context, scrapeFlag bool, scrapeQueue chan models.Regulation) {
	if scrapeFlag {
		go func() {
			err := scraper.FetchNewRegulations(ctx, scrapeQueue)
			if err != nil {
				logger.Error("Error in scraper: %v", err)
			}
		}()
	} else {
		close(scrapeQueue)
	}
}
