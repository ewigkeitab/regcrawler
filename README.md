# RegCrawler - 法規動態爬蟲與 AI 摘要工具

這是一個自動化的最新法規整理工具，爬找全國法規資料庫最新的法規資訊，並透過 Google Gemini AI API生成結構化且易讀的重點摘要。同時，此工具也提供多種輸出格式，包括 Markdown、JSON、Terminal 直接顯示等選項。

## 功能特色

- **自動化爬蟲**: 自動抓取最新的法規公告。
- **AI 摘要**: 整合 Google Gemini AI，自動提取法規重點，包括：
  - 法規類別與依據
  - 主要變革內容
  - 影響對象
  - 生效日期
- **高效能並發處理**: 採用 Concurrency (Goroutines & Channels) 架構。
- **多樣化輸出格式**:
  - `markdown`: 生成 Markdown 格式文件 (預設)。
  - `json`:  JSON 格式資料。
  - `mdstdout`: 直接在Terminal顯示Markdown格式 (支援顏色與樣式)。
- **AI 模型彈性**:
  - **模型選擇**: 可指定使用的 Gemini 模型 (如 `gemini-2.5-flash`, `gemini-2.0-flash` 等)。
  - **自訂 Prompt**: 支援載入外部的摘要 Prompt 提示詞調整。


## 快速開始

### 環境需求

- **Go**: 版本 1.25 或以上。
- **API Key**: Google Gemini API Key。 [LINK](https://aistudio.google.com/app/apikey)

### 安裝與建置

1. **下載**:
   ```bash
   git clone <repository-url>
   cd regcrawler
   ```

2. **安裝程式庫**:
   ```bash
   go mod tidy
   ```

3. **建置執行檔**:
   使用 Makefile 建置：
   ```bash
   make build  
   make release 
   ```
   或手動建置：
   ```bash
   go build -o regcrawler ./cmd/regcrawler
   ```

4. **設定環境變數**:
   將 Gemini API Key 設定到環境變數 `GEMINI_API_KEY` 中。
   ```bash
   export GEMINI_API_KEY='您的_API_KEY'
   ```

## 使用說明

### 基本指令

```bash
./regcrawler [選項]
```

### 命令參數 (Flags)

| 參數 | 預設值 | 說明 |
| :--- | :--- | :--- |
| `-format` | `markdown` | 輸出格式。可選值: `markdown`, `json`, `mdstdout`。 |
| `-model` | `gemini-2.5-flash` | 指定使用的 AI 模型版本。 |
| `-prompt` | 無 | 指定自訂 Prompt 文字檔的路徑。若未指定則使用內建預設值。 |
| `-output` | 無 | 指定輸出檔案名稱。若未指定則根據格式使用預設檔名。 |

### 使用範例

**1. 標準執行**
執行爬蟲並使用 AI 產生摘要，最後輸出為 `regulatory_report.md`。
```bash
./regcrawler
```

**2. 在終端機直接閱讀 (漂亮排版)**
不產生檔案，直接在終端機以彩色 Markdown 格式顯示結果。
```bash
./regcrawler -format=mdstdout
```

**3. 輸出為 JSON 檔**
適合需進一步資料處理的情境，輸出為 `processed_regulations.json`。
```bash
./regcrawler -format=json
```

**4. 更改 AI 模型**
切換使用 `gemini-2.5-flash`。
```bash
./regcrawler -model=gemini-2.5-flash
```
目前可用的常用模型包括：
- `gemini-2.5-flash`
- `gemini-2.5-flash-lite`
- `gemini-2.0-flash`
- `gemini-1.5-pro`

可以在 [Google AI Studio](https://ai.google.dev/gemini-api/docs/models?hl=zh-tw#model-versions) 查看完整的模型列表。

**5. 使用自訂 Prompt**
如果想要使用自己的提示詞內容，可以建立一個文字檔 (例如 `myprompt.txt`)，內容必須包含 `%s` (用來填入法規內文)。
```bash
./regcrawler -prompt=myprompt.txt
```

## 專案結構

- `cmd/regcrawler/`: 程式進入點 (main)。
- `pkg/scraper/`: 負責從網站爬取法規資料。
- `pkg/processor/`: 負責呼叫 AI 進行處理與生成摘要。
- `pkg/exporter/`: 負責將處理後的資料匯出成不同格式 (JSON, Markdown)。
- `pkg/logger/`: 提供美觀的終端機日誌輸出工具。
- `pkg/models/`: 定義資料結構。
- `prompt.txt`: 預設的 Prompt 範例。

