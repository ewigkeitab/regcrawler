package main

import (
	"context"
	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
	"strings"
	"time"
)

func runProcessor(ctx context.Context, cancel context.CancelFunc, processFlag bool, apiKey string, models []ModelList, interval time.Duration, promptTemplate string, scrapeQueue <-chan models.Regulation, resultQueue chan<- models.Regulation) {
	go func() {
		defer close(resultQueue)
		if processFlag {
			modelNames := make([]string, len(models))
			for i, m := range models {
				modelNames[i] = m.Name
			}

			err := processor.ProcessStream(ctx, apiKey, modelNames, promptTemplate, interval, scrapeQueue, resultQueue)
			if err != nil {
				logger.Error("Error in processor: %v", err)
				if strings.Contains(err.Error(), "429") {
					logger.Warn("All models rate limited, triggering graceful shutdown...")
					cancel()
				}
			}
		} else {
			for reg := range scrapeQueue {
				resultQueue <- reg
			}
		}
	}()
}
