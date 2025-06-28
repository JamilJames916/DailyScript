package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type WebScraper struct {
	client    *http.Client
	userAgent string
	delay     time.Duration
}

type ScrapedData struct {
	URL       string            `json:"url"`
	Title     string            `json:"title"`
	Links     []string          `json:"links"`
	Images    []string          `json:"images"`
	Text      string            `json:"text"`
	Emails    []string          `json:"emails"`
	Phones    []string          `json:"phones"`
	Timestamp time.Time         `json:"timestamp"`
}

func NewWebScraper() *WebScraper {
	return &WebScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "Go-Web-Scraper/1.0",
		delay:     1 * time.Second,
	}
}

func (ws *WebScraper) SetUserAgent(userAgent string) {
	ws.userAgent = userAgent
}

func (ws *WebScraper) SetDelay(delay time.Duration) {
	ws.delay = delay
}

func (ws *WebScraper) ScrapeURL(targetURL string) (*ScrapedData, error) {
	// Add delay to be respectful
	time.Sleep(ws.delay)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", ws.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	html := string(body)

	data := &ScrapedData{
		URL:       targetURL,
		Links:     ws.extractLinks(html, targetURL),
		Images:    ws.extractImages(html, targetURL),
		Title:     ws.extractTitle(html),
		Text:      ws.extractText(html),
		Emails:    ws.ExtractEmails(html),
		Phones:    ws.ExtractPhones(html),
		Timestamp: time.Now(),
	}

	return data, nil
}

func (ws *WebScraper) extractTitle(html string) string {
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (ws *WebScraper) extractLinks(html, baseURL string) []string {
	linkRegex := regexp.MustCompile(`(?i)<a[^>]*href=["']([^"']+)["'][^>]*>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)
	
	var links []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			link := ws.makeAbsolute(match[1], baseURL)
			if !seen[link] {
				links = append(links, link)
				seen[link] = true
			}
		}
	}
	
	return links
}

func (ws *WebScraper) extractImages(html, baseURL string) []string {
	imgRegex := regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']+)["'][^>]*>`)
	matches := imgRegex.FindAllStringSubmatch(html, -1)
	
	var images []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			img := ws.makeAbsolute(match[1], baseURL)
			if !seen[img] {
				images = append(images, img)
				seen[img] = true
			}
		}
	}
	
	return images
}

func (ws *WebScraper) extractText(html string) string {
	// Remove script and style tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")
	
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")
	
	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")
	
	// Clean up whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

func (ws *WebScraper) makeAbsolute(href, baseURL string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	relative, err := url.Parse(href)
	if err != nil {
		return href
	}

	return base.ResolveReference(relative).String()
}

func (ws *WebScraper) ExtractEmails(text string) []string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(text, -1)
	
	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, email := range emails {
		if !seen[email] {
			unique = append(unique, email)
			seen[email] = true
		}
	}
	
	return unique
}

func (ws *WebScraper) ExtractPhones(text string) []string {
	phoneRegex := regexp.MustCompile(`(\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	phones := phoneRegex.FindAllString(text, -1)
	
	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, phone := range phones {
		if !seen[phone] {
			unique = append(unique, phone)
			seen[phone] = true
		}
	}
	
	return unique
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run web-scraper.go <command> <url>")
		fmt.Println("Commands:")
		fmt.Println("  scrape <url>    - Full scrape of the webpage")
		fmt.Println("  links <url>     - Extract all links")
		fmt.Println("  images <url>    - Extract all images")
		fmt.Println("  emails <url>    - Extract email addresses")
		fmt.Println("  phones <url>    - Extract phone numbers")
		fmt.Println("  title <url>     - Extract page title")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run web-scraper.go scrape https://example.com")
		fmt.Println("  go run web-scraper.go links https://news.ycombinator.com")
		os.Exit(1)
	}

	command := os.Args[1]
	targetURL := os.Args[2]

	scraper := NewWebScraper()

	switch command {
	case "scrape":
		data, err := scraper.ScrapeURL(targetURL)
		if err != nil {
			log.Fatalf("Scraping failed: %v", err)
		}

		fmt.Printf("Title: %s\n", data.Title)
		fmt.Printf("URL: %s\n", data.URL)
		fmt.Printf("Scraped at: %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Found %d links\n", len(data.Links))
		fmt.Printf("Found %d images\n", len(data.Images))
		fmt.Printf("Found %d emails\n", len(data.Emails))
		fmt.Printf("Found %d phone numbers\n", len(data.Phones))

		if data.Text != "" {
			fmt.Println("\nText content preview:")
			preview := data.Text
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			fmt.Println(preview)
		}

	case "links":
		data, err := scraper.ScrapeURL(targetURL)
		if err != nil {
			log.Fatalf("Failed to scrape: %v", err)
		}

		fmt.Printf("Found %d links:\n", len(data.Links))
		for i, link := range data.Links {
			fmt.Printf("%d. %s\n", i+1, link)
		}

	case "images":
		data, err := scraper.ScrapeURL(targetURL)
		if err != nil {
			log.Fatalf("Failed to scrape: %v", err)
		}

		fmt.Printf("Found %d images:\n", len(data.Images))
		for i, image := range data.Images {
			fmt.Printf("%d. %s\n", i+1, image)
		}

	case "emails":
		data, err := scraper.ScrapeURL(targetURL)
		if err != nil {
			log.Fatalf("Scraping failed: %v", err)
		}

		fmt.Printf("Found %d email addresses:\n", len(data.Emails))
		for i, email := range data.Emails {
			fmt.Printf("%d. %s\n", i+1, email)
		}

	case "phones":
		data, err := scraper.ScrapeURL(targetURL)
		if err != nil {
			log.Fatalf("Scraping failed: %v", err)
		}

		fmt.Printf("Found %d phone numbers:\n", len(data.Phones))
		for i, phone := range data.Phones {
			fmt.Printf("%d. %s\n", i+1, phone)
		}

	case "title":
		data, err := scraper.ScrapeURL(targetURL)
		if err != nil {
			log.Fatalf("Scraping failed: %v", err)
		}

		fmt.Printf("Title: %s\n", data.Title)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
