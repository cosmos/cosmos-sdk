package module

import (
	"math/rand"
	"strings"
	"testing"
	"unicode"
)

func TestRandomWords(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	count := 5
	separator := "-"
	words, err := randomWords(r, count, separator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wordList := strings.Split(words, separator)
	if len(wordList) != count {
		t.Fatalf("expected %d words, got %d", count, len(wordList))
	}

	for _, word := range wordList {
		if len(word) < 4 || len(word) > 8 {
			t.Errorf("word %q has invalid length", word)
		}
		if unicode.IsUpper(rune(word[0])) {
			t.Errorf("word %q starts with an uppercase letter", word)
		}
	}
}

func TestRandomWordsDeterministic(t *testing.T) {
	seed := int64(42)
	r1 := rand.New(rand.NewSource(seed))
	r2 := rand.New(rand.NewSource(seed))
	count := 5
	separator := "-"

	words1, err := randomWords(r1, count, separator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	words2, err := randomWords(r2, count, separator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if words1 != words2 {
		t.Errorf("expected words to be the same, got %q and %q", words1, words2)
	}
}
