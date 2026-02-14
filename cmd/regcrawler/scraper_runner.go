package main

import (
	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/scraper"
)

func runScraper(scrapeFlag bool, scrapeQueue chan models.Regulation) {
	if scrapeFlag {
		go func() {
			err := scraper.FetchNewRegulations(scrapeQueue)
			if err != nil {
				logger.Error("Error in scraper: %v", err)
			}
		}()
	} else {
		close(scrapeQueue)
	}
}
