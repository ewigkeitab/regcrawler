package processor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"

	"github.com/charmbracelet/glamour"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const DefaultPrompt = `
你是一位法律專家。請分析以下法規文本，並提取繁體中文的重點。
請以 Markdown 格式提供結構化的摘要。
除非必要，只能輸出台灣的正體中文。
如果提供的資料不完整，請你搜尋網路補充相關資訊。
Text:
%s 
輸出格式:
- **標題**: [Brief Title in Traditional Chinese]
- **法規類別**: [Category of the regulation in Traditional Chinese]
- **法規依據**: [Legal basis of the regulation in Traditional Chinese]
- **狀態**: [Status of the regulation in Traditional Chinese]
- **主要變革**: [List of  main changes in Traditional Chinese]
- **影響對象**: [Who is affected in Traditional Chinese]
- **生效日期**: [Date if mentioned]
`

// ProcessStream reads regulations from input channel, processes them with Gemini, and sends to output channel
func ProcessStream(ctx context.Context, apiKey string, modelName string, promptTemplate string, input <-chan models.Regulation, output chan<- models.Regulation) error {
	defer close(output)

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("failed to create genai client: %w", err)
	}
	defer client.Close()

	logger.Info("Using AI Model: %s", modelName)
	model := client.GenerativeModel(modelName)

	processedCount := 0

	for reg := range input {
		// Basic check if we should process this regulation
		// In a real scenario, we might want to check against an existing list here or before sending to channel.
		// For this refactor, we process everything that comes in.

		logger.Section("Processing: " + reg.Title)

		if reg.Content == "" {
			logger.Muted("Skipping: No content found.")
			reg.Keypoints = "No content available to summarize."
			output <- reg
			continue
		}

		if reg.Keypoints != "" && !strings.HasPrefix(reg.Keypoints, "Error") && !strings.HasPrefix(reg.Keypoints, "⚠️") {
			if !strings.Contains(reg.Keypoints, "Affected Entities") {
				logger.Muted("Skipping: Already processed.")
				output <- reg
				continue
			}
		}

		prompt := fmt.Sprintf(promptTemplate, reg.Content)

		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			logger.Error("Error processing item: %v", err)
			reg.Keypoints = fmt.Sprintf("⚠️ Summary unavailable: Error (%v)", err)
			if strings.Contains(err.Error(), "429") {
				logger.Warn("Rate limit exceeded. Stopping stream.")
				output <- reg
				break
			}
		} else {
			text := ""
			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					text = string(txt)
				}
			}

			if text != "" {
				reg.Keypoints = text
				processedCount++
				logger.Success("Generated summary.")
			} else {
				reg.Keypoints = "⚠️ Summary unavailable: Empty response from API."
				logger.Warn("Empty response from AI.")
			}
		}

		output <- reg
		time.Sleep(5 * time.Second) // Rate limiting
	}

	return nil
}

func GenerateMarkdownReport(data []models.Regulation, filename string) {
	if filename == "" {
		filename = "regulatory_report.md"
	}
	logger.Info("Generating Markdown report: %s...", filename)

	content := GenerateMarkdown(data)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		logger.Error("Error saving report file: %v", err)
		return
	}
	logger.Success("Markdown report saved to %s", filename)
}

func GenerateMarkdown(data []models.Regulation) string {
	var sb strings.Builder
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	sb.WriteString(fmt.Sprintf("# 最新法規動態彙整 (Regulatory Update Summary)\nGenerated on: %s\n\n", timestamp))

	for _, item := range data {
		title := item.Title
		if title == "" {
			title = "無標題"
		}
		date := item.Date
		if date == "" {
			date = "未知日期"
		}
		link := item.Link
		if link == "" {
			link = "#"
		}
		keypoints := item.Keypoints
		if keypoints == "" {
			keypoints = "無摘要"
		}

		sb.WriteString(fmt.Sprintf("## [%s](%s)\n", title, link))
		sb.WriteString(fmt.Sprintf("**發布日期**: %s\n\n", date))

		if strings.HasPrefix(keypoints, "⚠️") {
			sb.WriteString(fmt.Sprintf("> [!WARNING]\n> %s\n", keypoints))
			sb.WriteString(fmt.Sprintf("> \n> [Original Text Link](%s)\n", link))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", keypoints))
		}

		sb.WriteString("\n---\n\n")
	}
	return sb.String()
}

func RenderMarkdown(data []models.Regulation) {
	content := GenerateMarkdown(data)

	out, err := glamour.Render(content, "dark")
	if err != nil {
		logger.Error("Error rendering markdown: %v", err)
		fmt.Println(content)
		return
	}
	fmt.Println(out)
}
