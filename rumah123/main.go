// search.go
package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
)

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

type SearchResult struct {
	Title              string
	Content            string
	URL                string
	Score              float64
	HighlightedContent template.HTML
}

func cleanContent(content string) string {
	// 1. Remove standard unwanted texts
	unwantedTexts := []string{
		"Baca juga:",
		"Baca Juga:",
		"KOMPAS.com -",
		"Simak breaking news",
		"Google News",
		"Terus ikuti",
		"Lebih banyak informasi",
	}

	for _, text := range unwantedTexts {
		content = strings.ReplaceAll(content, text, "")
	}

	// 2. Remove URL patterns
	urlPattern := regexp.MustCompile(`https?://[^\s)]+`)
	content = urlPattern.ReplaceAllString(content, "")

	// 3. Remove repeated spaces and clean up
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	return content
}

func getContentPreview(content, query string, maxLength int) string {
	cleanedContent := cleanContent(content)

	if len(cleanedContent) <= maxLength {
		return cleanedContent
	}

	lowerContent := strings.ToLower(cleanedContent)
	lowerQuery := strings.ToLower(query)

	pos := strings.Index(lowerContent, lowerQuery)
	if pos == -1 {
		return cleanedContent[:maxLength] + "..."
	}

	start := pos - maxLength/3
	if start < 0 {
		start = 0
	}

	end := start + maxLength
	if end > len(cleanedContent) {
		end = len(cleanedContent)
	}

	result := cleanedContent[start:end]
	if start > 0 {
		result = "..." + result
	}
	if end < len(cleanedContent) {
		result = result + "..."
	}

	return result
}

func highlightText(text string, query string) string {
	if query == "" {
		return text
	}
	words := strings.Fields(strings.ToLower(query))
	highlighted := text

	for _, word := range words {
		if len(word) < 2 {
			continue
		}
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(word))
		highlighted = re.ReplaceAllString(highlighted, `<em>$0</em>`)
	}

	return highlighted
}

func loadArticles() ([]Article, error) {
	var allArticles []Article

	// Define JSON files to read
	jsonFiles := []string{
		"propertiterkini/articles2.json",
		"propertyandthecity/articles.json",
		"rumah123/articles3.json",
	}

	for _, file := range jsonFiles {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		var articles []Article
		if err := json.Unmarshal(data, &articles); err != nil {
			log.Printf("Error parsing JSON from %s: %v", file, err)
			continue
		}

		allArticles = append(allArticles, articles...)
	}

	return allArticles, nil
}

func searching(query string, method string) []SearchResult {
	var results []SearchResult

	articles, err := loadArticles()
	if err != nil {
		log.Printf("Error loading articles: %v", err)
		return results
	}

	queryTokens := tokenize(query)

	for _, article := range articles {
		fullText := article.Title + " " + article.Content
		docTokens := tokenize(fullText)

		var score float64
		switch method {
		case "cosine":
			score = cosineSimilarity(queryTokens, docTokens)
		case "jaccard":
			score = jaccardSimilarity(queryTokens, docTokens)
		default:
			score = cosineSimilarity(queryTokens, docTokens)
		}

		if score > 0 {
			contentPreview := getContentPreview(article.Content, query, 300)
			highlightedContent := highlightText(contentPreview, query)

			results = append(results, SearchResult{
				Title:              article.Title,
				Content:            contentPreview,
				URL:                article.URL,
				Score:              score,
				HighlightedContent: template.HTML(highlightedContent),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	text = strings.Map(func(r rune) rune {
		if strings.ContainsRune(".,!?()[]{}:;\"", r) {
			return ' '
		}
		return r
	}, text)

	tokens := strings.Fields(text)

	stopWords := map[string]bool{
		"yang": true, "di": true, "ke": true, "dari": true,
		"dan": true, "atau": true, "ini": true, "itu": true,
		"juga": true, "sudah": true, "saya": true, "anda": true,
		"dia": true, "mereka": true, "kita": true, "akan": true,
		"bisa": true, "ada": true, "tidak": true, "saat": true,
		"oleh": true, "setelah": true, "para": true, "dengan": true,
	}

	filteredTokens := make([]string, 0)
	for _, token := range tokens {
		if !stopWords[token] && len(token) > 1 {
			filteredTokens = append(filteredTokens, token)
		}
	}

	return filteredTokens
}

func cosineSimilarity(queryTokens, docTokens []string) float64 {
	queryFreq := make(map[string]float64)
	docFreq := make(map[string]float64)

	for _, token := range queryTokens {
		queryFreq[token]++
	}

	for _, token := range docTokens {
		docFreq[token]++
	}

	var dotProduct, queryMagnitude, docMagnitude float64

	for token, freq := range queryFreq {
		if docFrequency, exists := docFreq[token]; exists {
			dotProduct += freq * docFrequency
		}
		queryMagnitude += freq * freq
	}

	for _, freq := range docFreq {
		docMagnitude += freq * freq
	}

	if queryMagnitude == 0 || docMagnitude == 0 {
		return 0
	}

	queryMagnitude = math.Sqrt(queryMagnitude)
	docMagnitude = math.Sqrt(docMagnitude)

	return dotProduct / (queryMagnitude * docMagnitude)
}

func jaccardSimilarity(queryTokens, docTokens []string) float64 {
	querySet := make(map[string]bool)
	docSet := make(map[string]bool)

	for _, token := range queryTokens {
		querySet[token] = true
	}

	for _, token := range docTokens {
		docSet[token] = true
	}

	intersection := 0
	for token := range querySet {
		if docSet[token] {
			intersection++
		}
	}

	union := len(querySet) + len(docSet) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}
