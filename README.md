# Information Retrieval Search Engine

A search engine implementation for Information Retrieval course project. This search engine implements text processing, indexing, and two similarity methods (Cosine and Jaccard) to provide relevant search results.

## Features

- Text Processing steps:
  1. Remove punctuations and numbers
  2. Remove Stopwords (Indonesian)
  3. Case folding
  4. Stemming (Indonesian)
  5. Tokenization

- Indexing & Search:
  - Inverted Index implementation
  - TF-IDF Weighting
  - Two similarity methods:
    - Cosine Similarity
    - Jaccard Similarity

- Web Interface:
  - Clean and responsive design
  - Shows top 10 relevant results per page
  - Result highlighting
  - Favicon support for different sources

## Screenshots

### Search Page
Here's the main search page where users can input their queries:

![241114_10h16m54s_screenshot](https://github.com/user-attachments/assets/97e55600-fcb0-4736-bff1-48b43c3a9fb5)

### Results Page
The results page showing the ranked documents:

![241114_10h18m58s_screenshot](https://github.com/user-attachments/assets/96ce83b4-4078-4bb5-99a5-858e7e086228)


## Implementation Details

### Text Processing
The implementation follows specific steps to process both queries and documents:
```go
func (tp *TextProcessor) ProcessText(text string) []string {
    // 1. Remove punctuations dan nomor/angka
    cleaned := tp.removePunctuationsAndNumbers(text)
    
    // 2. Remove Stopword
    withoutStopwords := tp.removeStopwords(cleaned)
    
    // 3. Case folding
    folded := tp.caseFolding(withoutStopwords)
    
    // 4. Stemming
    stemmed := tp.stemming(folded)
    
    // 5. Tokenisasi
    return stemmed
}
```

### Indexing
Uses inverted index structure for efficient searching:
```go
type InvertedIndex struct {
    Index map[string]*PostingList
}

type PostingList struct {
    DocFrequency int
    Postings     map[int]*Posting
}
```

### Similarity Methods

1. Cosine Similarity with TF-IDF:
   - Measures similarity based on the cosine of the angle between document vectors
   - Considers term frequency and inverse document frequency

2. Jaccard Similarity:
   - Measures similarity based on intersection over union of terms
   - Good for comparing document similarity regardless of size

## Project Structure

```
.
├── main.go             # Main application entry
├── search.go           # Core search implementation
├── templates/          # HTML templates
│   ├── index.html      # Search page template
│   └── results.html    # Results page template
└── static/             # Static assets
    ├── css/            # Stylesheets
    └── favicon/        # Favicon assets
```

## Setup and Running

1. Clone the repository
```bash
git clone github.com/Mahathirrr/search-engine2
```

2. Install dependencies
```bash
go mod tidy
```

3. Run the application
```bash
go run main.go
```

4. Open in browser
```
http://localhost:8080
```

## Dependencies

- Go 1.18+
- Gin Web Framework
- HTML Template
- Other standard Go libraries

## Performance

The search engine implements efficient indexing and searching mechanisms:
- Pre-processes and indexes documents
- Uses TF-IDF weighting for better relevance
- Provides fast search results through inverted index
- Supports pagination for large result sets
