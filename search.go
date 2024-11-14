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

// Struktur dasar
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
	Favicon            string
}

// Struktur untuk inverted index
type InvertedIndex struct {
	Index map[string]*PostingList
}

type PostingList struct {
	DocFrequency int
	Postings     map[int]*Posting
}

type Posting struct {
	DocID     int
	Frequency int
	Positions []int
}

// Text Processor
type TextProcessor struct {
	stopWords   map[string]bool
	punctuation *regexp.Regexp
	numbers     *regexp.Regexp
}

// Variabel global
var (
	prefixes = []string{
		"me", "pe", "be", "te", "di", "ke", "se",
		"ber", "per", "ter", "mem", "pem", "pen",
		"meng", "peng", "meny", "peny",
	}
	suffixes = []string{
		"kan", "an", "i", "lah", "kah", "nya", "ku", "mu",
		"wan", "wati", "isme",
	}
	textProcessor *TextProcessor
)

func init() {
	textProcessor = NewTextProcessor()
}

func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		stopWords:   initializeStopWords(),
		punctuation: regexp.MustCompile(`[^\w\s]`),
		numbers:     regexp.MustCompile(`\b\d+\b`),
	}
}

func initializeStopWords() map[string]bool {
	return map[string]bool{
		// Kata hubung
		"yang": true, "dan": true, "atau": true, "tetapi": true, "namun": true,
		"melainkan": true, "sedangkan": true, "sebaliknya": true,

		// Kata depan
		"di": true, "ke": true, "dari": true, "dalam": true, "kepada": true,
		"pada": true, "oleh": true, "untuk": true, "bagi": true, "tentang": true,
		"menurut": true, "seperti": true, "sebagai": true,

		// Kata tunjuk
		"ini": true, "itu": true, "tersebut": true, "berikut": true,

		// Kata ganti orang
		"saya": true, "anda": true, "dia": true, "mereka": true, "kita": true,
		"kami": true, "kamu": true, "ia": true, "beliau": true,

		// Kata bantu
		"akan": true, "sudah": true, "telah": true, "sedang": true, "masih": true,
		"hendak": true, "bisa": true, "dapat": true, "bukan": true, "jangan": true,

		// Kata keterangan
		"sangat": true, "hanya": true, "juga": true, "saja": true, "lagi": true,
		"sekarang": true, "yakni": true, "yaitu": true,

		// Kata tanya
		"apa": true, "siapa": true, "dimana": true, "kapan": true, "kenapa": true,
		"bagaimana": true, "mengapa": true,

		// Kata bilangan
		"satu": true, "dua": true, "tiga": true, "empat": true, "lima": true,
		"enam": true, "tujuh": true, "delapan": true, "sembilan": true, "sepuluh": true,
		"pertama": true, "kedua": true, "ketiga": true, "keempat": true, "kelima": true,
	}
}

// Text Processing Steps
// 1. Remove punctuations dan nomor/angka
func (tp *TextProcessor) removePunctuationsAndNumbers(text string) string {
	text = tp.punctuation.ReplaceAllString(text, " ")
	text = tp.numbers.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// 2. Remove Stopword
func (tp *TextProcessor) removeStopwords(text string) []string {
	words := strings.Fields(text)
	filtered := make([]string, 0)
	for _, word := range words {
		if !tp.stopWords[strings.ToLower(word)] {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// 3. Case folding
func (tp *TextProcessor) caseFolding(tokens []string) []string {
	folded := make([]string, len(tokens))
	for i, token := range tokens {
		folded[i] = strings.ToLower(token)
	}
	return folded
}

// 4. Stemming
func (tp *TextProcessor) stem(word string) string {
	if len(word) < 4 {
		return word
	}

	origWord := word

	// Coba hapus suffix terlebih dahulu
	for _, suffix := range suffixes {
		if strings.HasSuffix(word, suffix) {
			word = strings.TrimSuffix(word, suffix)
			break
		}
	}

	// Kemudian hapus prefix
	for _, prefix := range prefixes {
		if strings.HasPrefix(word, prefix) {
			stemmed := strings.TrimPrefix(word, prefix)
			if len(stemmed) >= 4 {
				word = stemmed
				break
			}
		}
	}

	if len(word) < 3 {
		return origWord
	}

	return word
}

func (tp *TextProcessor) stemming(tokens []string) []string {
	stemmed := make([]string, len(tokens))
	for i, token := range tokens {
		stemmed[i] = tp.stem(token)
	}
	return stemmed
}

// 5. Tokenisasi (final)
func (tp *TextProcessor) tokenize(text string) []string {
	return strings.Fields(text)
}

// Proses text lengkap dengan urutan yang benar
func (tp *TextProcessor) ProcessText(text string) []string {
	// 1. Remove punctuations dan nomor/angka
	cleaned := tp.removePunctuationsAndNumbers(text)

	// 2. Remove Stopword
	withoutStopwords := tp.removeStopwords(cleaned)

	// 3. Case folding
	folded := tp.caseFolding(withoutStopwords)

	// 4. Stemming
	stemmed := tp.stemming(folded)

	// 5. Tokenisasi adalah hasil akhir dari proses stemming
	return stemmed
}

// Fungsi untuk membuat inverted index baru
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Index: make(map[string]*PostingList),
	}
}

// Fungsi untuk membangun inverted index
func buildInvertedIndex(articles []Article) *InvertedIndex {
	idx := NewInvertedIndex()

	for docID, article := range articles {
		tokens := textProcessor.ProcessText(article.Title + " " + article.Content)

		// Track position untuk setiap term
		for pos, token := range tokens {
			if _, exists := idx.Index[token]; !exists {
				idx.Index[token] = &PostingList{
					DocFrequency: 0,
					Postings:     make(map[int]*Posting),
				}
			}

			if _, exists := idx.Index[token].Postings[docID]; !exists {
				idx.Index[token].Postings[docID] = &Posting{
					DocID:     docID,
					Frequency: 0,
					Positions: make([]int, 0),
				}
				idx.Index[token].DocFrequency++
			}

			posting := idx.Index[token].Postings[docID]
			posting.Frequency++
			posting.Positions = append(posting.Positions, pos)
		}
	}

	return idx
}

// Menghitung TF-IDF dengan inverted index
func calculateTFIDF(invertedIndex *InvertedIndex, totalDocs int) map[string]map[int]float64 {
	tfidfScores := make(map[string]map[int]float64)

	for term, postingList := range invertedIndex.Index {
		tfidfScores[term] = make(map[int]float64)

		// Hitung IDF: log(Total Dokumen / Dokumen yang mengandung term)
		idf := math.Log(float64(totalDocs) / float64(postingList.DocFrequency))

		for docID, posting := range postingList.Postings {
			// TF * IDF
			tf := float64(posting.Frequency)
			tfidfScores[term][docID] = tf * idf
		}
	}

	return tfidfScores
}

// Normalisasi vector
func normalizeVector(vector map[string]float64) map[string]float64 {
	normalized := make(map[string]float64)
	var magnitude float64

	// Hitung magnitude
	for _, weight := range vector {
		magnitude += weight * weight
	}
	magnitude = math.Sqrt(magnitude)

	// Normalisasi
	if magnitude > 0 {
		for term, weight := range vector {
			normalized[term] = weight / magnitude
		}
	}

	return normalized
}

// Cosine Similarity dengan TF-IDF
func cosineSimilarityWithTFIDF(queryVector map[string]float64, tfidfScores map[string]map[int]float64, docID int) float64 {
	docVector := make(map[string]float64)

	// Buat vektor dokumen dari TF-IDF scores
	for term, scores := range tfidfScores {
		if score, exists := scores[docID]; exists {
			docVector[term] = score
		}
	}

	// Normalisasi kedua vektor
	normalizedQuery := normalizeVector(queryVector)
	normalizedDoc := normalizeVector(docVector)

	// Hitung dot product
	var dotProduct float64
	for term, queryWeight := range normalizedQuery {
		if docWeight, exists := normalizedDoc[term]; exists {
			dotProduct += queryWeight * docWeight
		}
	}

	return dotProduct
}

// Jaccard Similarity dengan TF-IDF
func jaccardSimilarityWithTFIDF(queryVector map[string]float64, tfidfScores map[string]map[int]float64, docID int) float64 {
	querySet := make(map[string]bool)
	docSet := make(map[string]bool)

	// Convert weighted vectors to sets
	for term := range queryVector {
		querySet[term] = true
	}

	for term, scores := range tfidfScores {
		if _, exists := scores[docID]; exists {
			docSet[term] = true
		}
	}

	// Calculate intersection
	intersection := 0
	for term := range querySet {
		if docSet[term] {
			intersection++
		}
	}

	// Calculate union
	union := len(querySet) + len(docSet) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// Content Preview Generator
func getContentPreview(content, query string, maxLength int) string {
	cleanedContent := cleanContent(content)
	maxLength = 160

	if len(cleanedContent) <= maxLength {
		return cleanedContent
	}

	processedQueryTokens := textProcessor.ProcessText(query)
	processedContentTokens := textProcessor.ProcessText(cleanedContent)

	queryText := strings.Join(processedQueryTokens, " ")
	contentText := strings.Join(processedContentTokens, " ")

	pos := strings.Index(strings.ToLower(contentText), strings.ToLower(queryText))
	if pos == -1 {
		return cleanedContent[:maxLength] + "..."
	}

	// Cari posisi kata di konten asli
	words := strings.Fields(cleanedContent)
	wordCount := len(strings.Fields(contentText[:pos]))

	// Hitung posisi karakter berdasarkan jumlah kata
	wordPos := 0
	for i := 0; i < wordCount && i < len(words); i++ {
		wordPos += len(words[i]) + 1
	}

	start := wordPos - 60
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

// Clean content for better processing
func cleanContent(content string) string {
	// 1. Remove unwanted texts
	unwantedTexts := []string{
		"Baca juga:", "Baca Juga:",
		"Simak breaking news", "Google News",
		"Terus ikuti", "Lebih banyak informasi",
		"Follow", "Instagram", "Twitter", "Facebook",
		"Bagikan:", "Share:", "Read more",
	}

	for _, text := range unwantedTexts {
		content = strings.ReplaceAll(content, text, "")
	}

	// 2. Remove all URLs
	urlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`https?://\S+`),          // http:// atau https://
		regexp.MustCompile(`www\.\S+`),              // www.
		regexp.MustCompile(`\S+\.(com|net|org)\S*`), // domain common
	}

	for _, pattern := range urlPatterns {
		content = pattern.ReplaceAllString(content, "")
	}

	// 3. Remove email addresses
	emailPattern := regexp.MustCompile(`\S+@\S+\.\S+`)
	content = emailPattern.ReplaceAllString(content, "")

	// 4. Remove social media handles
	socialPattern := regexp.MustCompile(`@\S+`)
	content = socialPattern.ReplaceAllString(content, "")

	// 5. Remove special characters dan punctuation
	specialCharsPattern := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	content = specialCharsPattern.ReplaceAllString(content, " ")

	// 6. Remove standalone numbers
	numberPattern := regexp.MustCompile(`\s+\d+\s+`)
	content = numberPattern.ReplaceAllString(content, " ")

	// 7. Remove extra whitespace
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// 8. Remove leading/trailing spaces
	content = strings.TrimSpace(content)

	// 9. Normalize spaces setelah tanda baca
	content = regexp.MustCompile(`\s*[.,!?;:]\s*`).ReplaceAllString(content, " ")

	// 10. Remove repeated words
	words := strings.Fields(content)
	uniqueWords := make([]string, 0)
	prev := ""
	for _, word := range words {
		if word != prev {
			uniqueWords = append(uniqueWords, word)
			prev = word
		}
	}
	content = strings.Join(uniqueWords, " ")

	// 11. Normalize whitespace final
	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// Highlight matched text
func highlightText(text string, query string) string {
	if query == "" {
		return text
	}

	queryTokens := textProcessor.ProcessText(query)
	highlighted := text

	for _, token := range queryTokens {
		if len(token) < 2 {
			continue
		}
		pattern := `(?i)\b[\wа-я]*` + regexp.QuoteMeta(token) + `[\wа-я]*\b`
		re := regexp.MustCompile(pattern)
		highlighted = re.ReplaceAllString(highlighted, `<em>$0</em>`)
	}

	return highlighted
}

// Get favicon path for URL
func getFaviconPath(url string) string {
	switch {
	case strings.HasPrefix(url, "https://artikel.rumah123.com/"):
		return "/static/rumah123.png"
	case strings.HasPrefix(url, "https://propertiterkini.com/"):
		return "/static/propertiterkini.png"
	case strings.HasPrefix(url, "https://propertyandthecity.com/"):
		return "/static/propertyandthecity.png"
	default:
		return "/static/favicon.svg"
	}
}

// Load articles from JSON file
func loadArticles() ([]Article, error) {
	var allArticles []Article

	data, err := ioutil.ReadFile("articles.json")
	if err != nil {
		log.Printf("Error reading articles.json: %v", err)
		return nil, err
	}

	if err := json.Unmarshal(data, &allArticles); err != nil {
		log.Printf("Error parsing JSON from articles.json: %v", err)
		return nil, err
	}

	return allArticles, nil
}

// Main search function
func searching(query string, method string) []SearchResult {
	articles, err := loadArticles()
	if err != nil {
		log.Printf("Error loading articles: %v", err)
		return nil
	}

	// Build inverted index
	invertedIndex := buildInvertedIndex(articles)

	// Calculate TF-IDF scores
	tfidfScores := calculateTFIDF(invertedIndex, len(articles))

	// Process query
	queryTokens := textProcessor.ProcessText(query)
	queryVector := make(map[string]float64)
	for _, token := range queryTokens {
		queryVector[token]++
	}

	var results []SearchResult

	for i, article := range articles {
		var score float64
		switch method {
		case "cosine":
			score = cosineSimilarityWithTFIDF(queryVector, tfidfScores, i)
		case "jaccard":
			score = jaccardSimilarityWithTFIDF(queryVector, tfidfScores, i)
		default:
			score = cosineSimilarityWithTFIDF(queryVector, tfidfScores, i)
		}

		if score > 0 {
			contentPreview := getContentPreview(article.Content, query, 160)
			highlightedContent := highlightText(contentPreview, query)

			results = append(results, SearchResult{
				Title:              article.Title,
				Content:            contentPreview,
				URL:                article.URL,
				Score:              score,
				HighlightedContent: template.HTML(highlightedContent),
				Favicon:            getFaviconPath(article.URL),
			})
		}
	}

	// Sort results by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
