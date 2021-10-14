package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestWordizer(t *testing.T) {
	text := strings.NewReader("   one two three's !!! fou!rr five SIX seVen's ei")
	words := &Wordizer{in: text, keep: "'.,"}
	fmt.Printf("text: %v\n", text)
	re := []string{}
	for {
		word, err := words.ReadWord()
		if err != nil && err != io.EOF {
			t.Fatalf("error; %v", err)
		}
		if word == "" {
			break
		}
		re = append(re, word)
	}
	if l := len(re); l != 9 {
		t.Errorf("wrong number of words: %d", l)
	}
	joined := strings.Join(re, "/")
	if joined != "one/two/three's/fou/rr/five/SIX/seVen's/ei" {
		t.Errorf("wrong words: %v", words)
	}
	fmt.Printf("words: %v\n", re)
}

func TestWordizerEmpty(t *testing.T) {
	words := &Wordizer{in: strings.NewReader(""), keep: "'.,"}
	_, err := words.ReadWord()
	if err != io.EOF {
		t.Errorf("bad error: %v", err)
	}
}

type fakewords struct {
	words []string
}

func (fw *fakewords) ReadWord() (string, error) {
	if len(fw.words) == 0 {
		return "", io.EOF
	}
	w := fw.words[0]
	fw.words = fw.words[1:]
	return w, nil
}

func TestTrigramizer(t *testing.T) {
	fw := &fakewords{words: []string{"1", "2", "3", "4", "5", "6"}}
	tz := &Trigramizer{in: fw}
	re, err := tz.ReadTrigram()
	if err != nil {
		t.Error("error")
	}
	if s := re.String(); s != "1 2 3" {
		t.Errorf("bad 1: %s", s)
	}
	re, err = tz.ReadTrigram()
	if s := re.String(); s != "2 3 4" {
		t.Errorf("bad 1: %s", s)
	}
}

func TestTrigramizerShort(t *testing.T) {
	fw := &fakewords{words: []string{"1", "2"}}
	tz := &Trigramizer{in: fw}
	_, err := tz.ReadTrigram()
	if err != io.EOF {
		t.Errorf("error: %v", err)
	}
}

func TestTrigramizerEmpty(t *testing.T) {
	fw := &fakewords{words: []string{}}
	tz := &Trigramizer{in: fw}
	_, err := tz.ReadTrigram()
	if err != io.EOF {
		t.Errorf("error: %v", err)
	}
}

func TestTrigrams(t *testing.T) {
	tg := NewTrigrams()

	in := []Trigram{
		{"a", "b", "c"},
		{"b", "c", "d"},
		{"c", "d", "e"},
	}
	tg.InputTrigrams(in)

	out, err := tg.GenerateN("a b", 100)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// only one output is possible from the trigrams
	if out != "a b c d e" {
		t.Errorf("bad output: %s", out)
	}
}
