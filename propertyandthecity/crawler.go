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
	Title   string    `json:"title"`
	Content string    `json:"content"`
	URL     string    `json:"url"`
	Date    time.Time `json:"date"`
	Author  string    `json:"author"`
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
		colly.AllowedDomains("propertyandthecity.com"),
		colly.MaxDepth(3),
		colly.Async(true),
	)

	// Create a slice to store all articles
	var articles []Article

	// Set up rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 3 * time.Second,
		Parallelism: 3,
	})

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "propertyandthecity.com") {
			fmt.Printf("%s[LINK] Found: %s%s\n", colorBlue, link, colorReset)
			e.Request.Visit(link)
		}
	})

	// Extract article data
	c.OnHTML("article", func(e *colly.HTMLElement) {
		article := Article{}

		// Extract title
		article.Title = strings.TrimSpace(e.ChildText("h1.entry-title"))

		// Extract and concatenate content from all p tags
		var contentParts []string
		e.ForEach("div.td-post-content p", func(_ int, el *colly.HTMLElement) {
			if text := strings.TrimSpace(el.Text); text != "" {
				contentParts = append(contentParts, text)
			}
		})
		// Join all content parts with newlines
		article.Content = strings.Join(contentParts, "\n")

		// Extract URL
		article.URL = e.Request.URL.String()

		// Extract date
		dateStr := e.ChildText("time.entry-date")
		if dateStr != "" {
			parsedDate, err := time.Parse("January 2, 2006", dateStr)
			if err == nil {
				article.Date = parsedDate
			}
		}

		// Extract author
		article.Author = strings.TrimSpace(e.ChildText(".td-post-author-name a"))

		if article.Title != "" && article.Content != "" {
			fmt.Printf("%s[ARTICLE] Successfully scraped: %s%s\n", colorGreen, article.Title, colorReset)
			fmt.Printf("%s[INFO] Author: %s | Date: %s%s\n", colorYellow, article.Author, article.Date.Format("2006-01-02"), colorReset)

			// Print content length for verification
			fmt.Printf("%s[INFO] Content length: %d characters%s\n", colorYellow, len(article.Content), colorReset)

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
	err := c.Visit("https://propertyandthecity.com")
	if err != nil {
		log.Fatal("Failed to start scraping:", err)
	}

	// Wait for all scraping jobs to complete
	c.Wait()

	// Save results to JSON file
	outputFile, err := os.Create("articles.json")
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
