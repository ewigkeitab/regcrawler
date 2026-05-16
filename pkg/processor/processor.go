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

// ProcessStream reads regulations from input channel, processes them with Gemini (falling back to next models on 429), and sends to output channel
func ProcessStream(ctx context.Context, apiKey string, modelNames []string, promptTemplate string, interval time.Duration, input <-chan models.Regulation, output chan<- models.Regulation) error {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("failed to create genai client: %w", err)
	}
	defer client.Close()

	if len(modelNames) == 0 {
		return fmt.Errorf("no models provided")
	}

	currentModelIdx := 0
	logger.Info("Starting processor with primary model: %s", modelNames[currentModelIdx])

	for reg := range input {
		select {
		case <-ctx.Done():
			logger.Warn("Processor stopping: context cancelled.")
			return ctx.Err()
		default:
		}

		logger.Section("Processing: " + reg.Title)

		if reg.Content == "" {
			logger.Muted("Skipping: No content found.")
			reg.Keypoints = "No content available to summarize."
			reg.Content = reg.Title
			output <- reg
			continue
		}

		// Skip if already processed (unless it's an error/warning)
		if reg.Keypoints != "" && !strings.HasPrefix(reg.Keypoints, "Error") && !strings.HasPrefix(reg.Keypoints, "[warning]") {
			if !strings.Contains(reg.Keypoints, "Affected Entities") {
				logger.Muted("Skipping: Already processed.")
				output <- reg
				continue
			}
		}

		prompt := fmt.Sprintf(promptTemplate, reg.Content)

		// Fallback loop for the current regulation
		var lastErr error
		for currentModelIdx < len(modelNames) {
			modelName := modelNames[currentModelIdx]
			model := client.GenerativeModel(modelName)

			resp, err := model.GenerateContent(ctx, genai.Text(prompt))
			if err != nil {
				lastErr = err
				if strings.Contains(err.Error(), "429") {
					logger.Warn("Model %s returned 429 (Too Many Requests).", modelName)
					currentModelIdx++
					if currentModelIdx < len(modelNames) {
						logger.Info("Falling back to next model: %s", modelNames[currentModelIdx])
						continue // Retry with next model
					} else {
						logger.Error("All models exhausted or rate limited.")
						reg.Keypoints = fmt.Sprintf("[warning] Summary unavailable: All models rate limited (Last error: %v)", err)
						break
					}
				} else {
					logger.Error("Error processing item with %s: %v", modelName, err)
					reg.Keypoints = fmt.Sprintf("[warning] Summary unavailable: Error (%v)", err)
					break // Non-429 error, don't fallback?
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
					logger.Success("Generated summary using %s.", modelName)
					lastErr = nil // Success
				} else {
					reg.Keypoints = "[warning] Summary unavailable: Empty response from API."
					logger.Warn("Empty response from AI (%s).", modelName)
				}
				break // Success or empty response, move to next item
			}
		}

		output <- reg

		if currentModelIdx >= len(modelNames) {
			logger.Error("Processor stopping: All models exhausted.")
			return fmt.Errorf("all models exhausted: %w", lastErr)
		}

		// Rate limiting with context awareness
		select {
		case <-time.After(interval):
			// Continue
		case <-ctx.Done():
			logger.Warn("Processor interrupted during rate limiting.")
			return ctx.Err()
		}
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
	sb.WriteString(fmt.Sprintf("# 最新法規動態彙整\n整理時間: %s\n\n", timestamp))

	for _, item := range data {
		sb.WriteString(GenerateItemMarkdown(item))
	}
	return sb.String()
}

func GenerateItemMarkdown(item models.Regulation) string {
	var sb strings.Builder
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

	if strings.HasPrefix(keypoints, "[warning]") {
		sb.WriteString(fmt.Sprintf("> [!WARNING]\n> %s\n", keypoints))
		sb.WriteString(fmt.Sprintf("> \n> [原始資料來源](%s)\n", link))
	} else {
		sb.WriteString(fmt.Sprintf("%s\n", keypoints))
	}

	sb.WriteString("\n---\n\n")
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
