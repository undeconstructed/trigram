package main

import (
	"fmt"
	"testing"
)

func TestNextWord(t *testing.T) {
	text := "   one two three's !!! fou!rr five SIX seVen ei"
	fmt.Printf("text: %v\n", text)
	words := []string{}
	for {
		word, rest := nextWord(text)
		if word == "" {
			break
		}
		words = append(words, word)
		text = rest
	}
	fmt.Printf("words: %v\n", words)
}
