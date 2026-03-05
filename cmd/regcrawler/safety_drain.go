package main

import (
	"context"
	"strings"
	"time"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"
	"regcrawler/pkg/storage"
)

func startSafetyDrain(ctx context.Context, processQueue <-chan models.Regulation) {
	go func() {
		<-ctx.Done()

		// Wait a small amount for the feeder to stop its DB operations
		time.Sleep(200 * time.Millisecond)

		logger.Warn("Draining process queue to database for safety...")
		count := 0
		for {
			select {
			case reg, ok := <-processQueue:
				if !ok {
					logDrain(count)
					return
				}
				if err := storage.SaveUnprocessed(reg); err != nil {
					if !strings.Contains(err.Error(), "locked") {
						logger.Error("Failed to save drained item: %v", err)
					}
				}
				count++
			default:
				logDrain(count)
				return
			}
		}
	}()
}

func logDrain(count int) {
	if count > 0 {
		logger.Info("Drained %d items back to database.", count)
	}
}
