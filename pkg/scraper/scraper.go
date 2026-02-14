package scraper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"regcrawler/pkg/logger"
	"regcrawler/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

const (
	BaseURL        = "https://law.moj.gov.tw"
	NewsURL        = "https://law.moj.gov.tw/News/NewsList.aspx"
	GazetteBaseURL = "https://gazette.nat.gov.tw"
	UserAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
)

func FetchNewRegulations(out chan<- models.Regulation) error {
	defer close(out)
	logger.Info("Fetching new regulations from %s...", NewsURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", NewsURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// Parsing the table
	count := 0
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 4 {
			dateROC := strings.TrimSpace(cells.Eq(1).Text())
			category := strings.TrimSpace(cells.Eq(2).Text())
			linkTag := cells.Eq(3).Find("a#hlkNAME")

			if linkTag.Length() > 0 {
				title := strings.TrimSpace(linkTag.Text())
				href, exists := linkTag.Attr("href")
				if exists {
					if strings.HasPrefix(href, "/") {
						href = BaseURL + href
					} else if !strings.HasPrefix(href, "http") {
						// Handle cases like "NewsDetail.aspx?..."
						href = BaseURL + "/News/" + href
					}

					// Date conversion
					dateAD := dateROC
					parts := strings.Split(dateROC, "-")
					if len(parts) == 3 {
						year, err := strconv.Atoi(parts[0])
						if err == nil {
							dateAD = fmt.Sprintf("%d-%s-%s", year+1911, parts[1], parts[2])
						}
					}

					content := FetchContent(href)
					count++

					// Log less verbosely, maybe just a muted log for each item found
					// logger.Muted("Found: %s", title)

					out <- models.Regulation{
						Title:    title,
						Date:     dateAD,
						Category: category,
						Link:     href,
						Content:  content,
					}
				}
			}
		}
	})

	logger.Success("Fetched %d regulations.", count)
	return nil
}

func FetchContent(pageURL string) string {
	// Reduce verbosity
	// logger.Muted("Fetching content from %s...", pageURL)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		logger.Error("Error creating request: %v", err)
		return ""
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error fetching content: %v", err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Error("Error parsing content: %v", err)
		return ""
	}

	var content string

	// Check for "Web Text Version" (Gazette specific)
	webTextLink := doc.Find("a[title*='網頁文字版']").First()
	if webTextLink.Length() == 0 {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "網頁文字版") {
				webTextLink = s
				return
			}
		})
	}

	if webTextLink.Length() > 0 {
		href, exists := webTextLink.Attr("href")
		if exists {
			targetURL := href
			if strings.HasPrefix(href, "/") {
				targetURL = GazetteBaseURL + href
			}
			// logger.Muted("Redirecting to Web Text Version...")

			respText, err := client.Get(targetURL)
			if err == nil {
				defer respText.Body.Close()
				docText, err := goquery.NewDocumentFromReader(respText.Body)
				if err == nil {
					content = docText.Text()
					content = strings.TrimSpace(content)
					return content
				}
			}
		}
	}

	// Fallback extraction
	contentDiv := doc.Find("div.Data_Info")
	if contentDiv.Length() == 0 {
		contentDiv = doc.Find("div.ContentPage")
	}
	if contentDiv.Length() == 0 {
		contentDiv = doc.Find("div.content")
	}

	if contentDiv.Length() > 0 {
		content = strings.TrimSpace(contentDiv.Text())
	} else {
		// logger.Warn("Could not identify content container for %s", pageURL)
	}

	return content
}
