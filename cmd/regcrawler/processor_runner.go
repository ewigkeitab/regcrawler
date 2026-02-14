package main

import (
	"context"
	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/processor"
)

func runProcessor(ctx context.Context, processFlag bool, apiKey, modelName, promptTemplate string, scrapeQueue <-chan models.Regulation, resultQueue chan<- models.Regulation) {
	go func() {
		if processFlag {
			err := processor.ProcessStream(ctx, apiKey, modelName, promptTemplate, scrapeQueue, resultQueue)
			if err != nil {
				logger.Error("Error in processor: %v", err)
			}
		} else {
			for reg := range scrapeQueue {
				resultQueue <- reg
			}
			close(resultQueue)
		}
	}()
}
