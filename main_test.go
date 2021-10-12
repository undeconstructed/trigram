package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestNextWord(t *testing.T) {
	text := strings.NewReader("   one two three's !!! fou!rr five SIX seVen's ei")
	fmt.Printf("text: %v\n", text)
	words := []string{}
	for {
		word := nextWord(text)
		if word == "" {
			break
		}
		words = append(words, word)
	}
	if l := len(words); l != 9 {
		t.Errorf("wrong number of words: %d", l)
	}
	fmt.Printf("words: %v\n", words)
}
