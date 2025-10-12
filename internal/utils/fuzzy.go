package utils

import (
	"strings"
	"unicode"
)

// FuzzyMatch performs a simple fuzzy search match
// Returns the match score (0 = no match, higher = better match)
func FuzzyMatch(pattern, text string) int {
	if pattern == "" {
		return 100 // Empty pattern matches everything with high score
	}

	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	patternRunes := []rune(pattern)
	textRunes := []rune(text)

	patternIndex := 0
	score := 0
	consecutiveMatches := 0

	for _, char := range textRunes {
		if patternIndex < len(patternRunes) && char == patternRunes[patternIndex] {
			score += 10 + consecutiveMatches*2 // Bonus for consecutive matches
			consecutiveMatches++
			patternIndex++
		} else {
			consecutiveMatches = 0
		}
	}

	// If we matched the entire pattern, return the score
	if patternIndex == len(patternRunes) {
		// Bonus for shorter text (prefer more concise matches)
		textLength := len(textRunes)
		if textLength > 0 {
			score += 100 / textLength
		}
		return score
	}

	return 0 // No match
}

// FuzzySearch performs fuzzy search on a slice of strings
// Returns matches sorted by relevance score
type SearchResult struct {
	Text  string
	Score int
}

func FuzzySearch(pattern string, texts []string) []SearchResult {
	var results []SearchResult

	for _, text := range texts {
		score := FuzzyMatch(pattern, text)
		if score > 0 {
			results = append(results, SearchResult{
				Text:  text,
				Score: score,
			})
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// SplitWords splits text into words for better search matching
func SplitWords(text string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	return words
}

// ContainsAnyWord checks if any of the search terms appear in the text
func ContainsAnyWord(searchTerms, textWords []string) bool {
	for _, search := range searchTerms {
		for _, word := range textWords {
			if strings.Contains(word, search) || strings.Contains(search, word) {
				return true
			}
		}
	}
	return false
}