# RegCrawler - Taiwan Regulatory Crawler & AI Summary Tool

This is an automated tool for organizing the latest regulations. It crawls the latest regulatory information from the Laws and Regulations Database of the Republic of China (Taiwan) and generates structured, readable key summaries through the Google Gemini AI API. At the same time, this tool also provides multiple output formats, including Markdown, JSON, and direct Terminal display options.

## Features

- **Automated Crawler**: Automatically fetches the latest regulatory announcements.
- **AI-Powered Summary**: Integrates Google Gemini AI to automatically extract key points, including:
  - Regulatory category and basis
  - Main changes
  - Affected parties
  - Effective date
- **High-Performance Concurrent Processing**: Built with Go's Concurrency (Goroutines & Channels) architecture.
- **Multiple Output Formats**:
  - `markdown`: Generates Markdown files (default).
  - `json`: Data in JSON format.
  - `mdstdout`: Displays Markdown directly in the terminal (with color and style support).
- **AI Model Flexibility**:
  - **Model Selection**: Specify the Gemini model (e.g., `gemini-2.5-flash`, `gemini-2.0-flash`).
  - **Custom Prompt**: Supports loading external Prompt templates to adjust the AI's summary style.


## Quick Start

### Prerequisites

- **Go**: Version 1.25 or higher.
- **API Key**: Google Gemini API Key. [Apply here](https://aistudio.google.com/app/apikey)

### Installation & Build

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   cd regcrawler
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build Binary**:
   Using Makefile:
   ```bash
   make build  
   make release 
   ```
   Or manually:
   ```bash
   go build -o regcrawler ./cmd/regcrawler
   ```

4. **Set Environment Variable**:
   Set your Gemini API Key in the `GEMINI_API_KEY` environment variable.
   ```bash
   export GEMINI_API_KEY='your_API_KEY'
   ```

## Usage

### Basic Command

```bash
./regcrawler [options]
```

### Command Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-format` | `markdown` | Output format. Options: `markdown`, `json`, `mdstdout`. |
| `-model` | `gemini-2.5-flash` | Specify the AI model version. |
| `-prompt` | None | Path to a custom Prompt text file. Uses internal default if not specified. |
| `-output` | None | Path to the output file. Uses internal default if not specified. |

### Examples

**1. Standard Execution**
Runs the crawler, uses AI for summaries, and outputs to `regulatory_report.md`.
```bash
./regcrawler
```

**2. Read Directly in Terminal (Beautiful Layout)**
Displays results in colored Markdown format directly in the terminal without generating a file.
```bash
./regcrawler -format=mdstdout
```

**3. Output to JSON File**
Suitable for further data processing. Outputs to `processed_regulations.json`.
```bash
./regcrawler -format=json
```

**4. Change AI Model**
Switch to `gemini-2.5-flash`.
```bash
./regcrawler -model=gemini-2.5-flash
```
Commonly used models include:
- `gemini-2.5-flash`
- `gemini-2.5-flash-lite`
- `gemini-2.0-flash`
- `gemini-1.5-pro`

You can view the full list of models on [Google AI Studio](https://ai.google.dev/gemini-api/docs/models?hl=en#model-versions).

**5. Use Custom Prompt**
To adjust the AI's summary format or tone, create a text file (e.g., `myprompt.txt`) containing `%s` (where the regulatory text will be inserted).
```bash
./regcrawler -prompt=myprompt.txt
```

## Project Structure

- `cmd/regcrawler/`: Entry point (main).
- `pkg/scraper/`: Responsible for crawling regulatory data from websites.
- `pkg/processor/`: Handles AI processing and summary generation.
- `pkg/exporter/`: Responsible for exporting results to different formats (JSON, Markdown).
- `pkg/logger/`: Provides beautiful terminal log output tools.
- `pkg/models/`: Defines data structures.
- `prompt.txt`: Default Prompt example.

