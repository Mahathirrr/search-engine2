package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Article represents the structure of our scraped data
type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// Terminal colors for better visibility
const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

func main() {
	// Initialize collector
	c := colly.NewCollector(
		colly.AllowedDomains("artikel.rumah123.com"),
		colly.MaxDepth(3),
		colly.Async(true),
	)

	// Create a slice to store all articles
	var articles []Article

	// Set up rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second,
		Parallelism: 4,
	})

	// Find and visit all links within the specified domain
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.HasPrefix(link, "https://artikel.rumah123.com/") {
			fmt.Printf("%s[LINK] Found: %s%s\n", colorBlue, link, colorReset)
			e.Request.Visit(link)
		}
	})

	// Extract article data
	c.OnHTML("article", func(e *colly.HTMLElement) {
		article := Article{}

		// Extract title
		article.Title = strings.TrimSpace(e.ChildText("h1.heading-3"))

		// Extract and concatenate content from all p tags
		var contentParts []string
		e.ForEach("div.content p", func(_ int, el *colly.HTMLElement) {
			if text := strings.TrimSpace(el.Text); text != "" {
				contentParts = append(contentParts, text)
			}
		})
		// Join all content parts with newlines
		article.Content = strings.Join(contentParts, "\n")

		// Extract URL
		article.URL = e.Request.URL.String()

		if article.Title != "" && article.Content != "" {
			fmt.Printf("%s[ARTICLE] Successfully scraped: %s%s\n", colorGreen, article.Title, colorReset)
			articles = append(articles, article)
		}
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("%s[ERROR] Failed to scrape %s: %s%s\n", colorRed, r.Request.URL, err, colorReset)
	})

	// Before making a request
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("%s[VISITING] %s%s\n", colorBlue, r.URL.String(), colorReset)
	})

	// Start scraping
	fmt.Println("🚀 Starting scraping process...")
	startTime := time.Now()
	err := c.Visit("https://artikel.rumah123.com/")
	if err != nil {
		log.Fatal("Failed to start scraping:", err)
	}

	// Wait for all scraping jobs to complete
	c.Wait()

	// Save results to JSON file
	outputFile, err := os.Create("articles3.json")
	if err != nil {
		log.Fatal("Failed to create output file:", err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(articles); err != nil {
		log.Fatal("Failed to encode articles to JSON:", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✨ Scraping completed in %s\n", duration)
	fmt.Printf("📦 Total articles scraped: %d\n", len(articles))
	fmt.Printf("💾 Results saved to articles.json\n")
}
