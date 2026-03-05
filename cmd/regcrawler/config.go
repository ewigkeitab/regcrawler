package main

import (
	"flag"
	"fmt"
	"os"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/processor"
)

// Config holds all startup configurations
type Config struct {
	ScrapeFlag     bool
	ProcessFlag    bool
	SkipAIFlag     bool
	FormatFlag     string
	ModelFlag      string
	PromptTemplate string
	OutputFlag     string
	APIKey         string
}

func parseConfig() *Config {
	cfg := &Config{}
	flag.BoolVar(&cfg.ScrapeFlag, "scrape", true, "Run the scraper to fetch new regulations")
	flag.BoolVar(&cfg.ProcessFlag, "process", true, "Run the AI processor to summarize regulations")
	flag.BoolVar(&cfg.SkipAIFlag, "skip-ai", false, "Skip AI processing and just generate report")
	flag.StringVar(&cfg.FormatFlag, "format", "markdown", "Output format: markdown or json")
	flag.StringVar(&cfg.ModelFlag, "model", "gemini-2.5-flash", "AI Model to use (e.g., gemini-2.0-flash, gemini-2.5-flash)")
	promptFlag := flag.String("prompt", "", "Path to custom prompt text file")
	flag.StringVar(&cfg.OutputFlag, "output", "", "Output file name (optional)")

	flag.Parse()

	if cfg.SkipAIFlag {
		cfg.ProcessFlag = false
	}

	// Load prompt
	cfg.PromptTemplate = processor.DefaultPrompt
	if *promptFlag != "" {
		content, err := os.ReadFile(*promptFlag)
		if err != nil {
			logger.Error("Error reading prompt file: %v", err)
			os.Exit(1)
		}
		cfg.PromptTemplate = string(content)
		logger.Info("Loaded custom prompt from %s", *promptFlag)
	}

	cfg.APIKey = os.Getenv("GEMINI_API_KEY")
	if cfg.ProcessFlag && cfg.APIKey == "" {
		logger.Error("GEMINI_API_KEY environment variable not set.")
		fmt.Println("Please set it using: export GEMINI_API_KEY='your_key_here'")
		os.Exit(1)
	}

	return cfg
}
