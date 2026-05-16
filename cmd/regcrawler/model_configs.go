package main

import (
	"regcrawler/pkg/logger"
	"time"
)

type ProcessorConfig struct {
	PreferredModel string
	IntervalTime   time.Duration
}
type ModelList struct {
	Provider string
	Name     string
}

var processorConfig = ProcessorConfig{
	PreferredModel: "gemini-3.1-pro-preview",
	IntervalTime:   time.Second * 20,
}

func GetProcessorConfig() ProcessorConfig {
	return processorConfig
}

func GetModelList() []ModelList {
	return []ModelList{
		{"gemini", "gemini-3.1-pro-preview"},
		{"gemini", "gemini-3.1-pro-flash"},
		{"gemini", "gemini-2.5-pro"},
		{"gemini", "gemini-2.5-flash"},
		{"gemini", "gemini-1.5-pro"},
		{"gemini", "gemini-1.5-flash"},
	}
}

func GetProviderByName(name string) string {
	for _, model := range GetModelList() {
		if model.Name == name {
			return model.Provider
		}
	}
	logger.Error("Model %s not found", name)
	return ""
}
