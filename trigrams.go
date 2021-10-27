package main

import (
	"errors"
	"io"
	"math/rand"
	"strings"
	"sync"
)

// WordReader has ReadWord
type WordReader interface {
	ReadWord() (string, error)
}

// Wordizer splits a stream into words
type Wordizer struct {
	keep string
	in   io.RuneReader
}

// Next acquired a word from the stream
func (w *Wordizer) ReadWord() (string, error) {
	word := ""
	for {
		rune, _, err := w.in.ReadRune()
		if err != nil {
			if err == io.EOF && len(word) > 0 {
				return word, nil
			}
			return "", err
		}
		if rune >= 'a' && rune <= 'z' {
			word = word + string(rune)
		} else if rune >= 'A' && rune <= 'Z' {
			// XXX - change case ?
			word = word + string(rune)
		} else if strings.Contains(w.keep, string(rune)) {
			word = word + string(rune)
		} else {
			if len(word) > 0 {
				// found a word, so consider this rune to be a break
				break
			}
			// else totally ignore this rune
		}
	}
	return word, nil
}

// Trigram is just three strings
type Trigram []string

func (t Trigram) String() string {
	ss := []string(t)
	return strings.Join(ss[:], " ")
}

func (t Trigram) Prefix(n int) Trigram {
	return Trigram(t[0:n])
}

func (t Trigram) Last() string {
	return t[len(t)-1]
}

// TrigramReader has ReadTrigram
type TrigramReader interface {
	TrigramWord() (Trigram, error)
}

// Trigramizer pulls successive Trigrams from a stream
type Trigramizer struct {
	// last n-1 words read
	last []string
	n    int
	in   WordReader
}

// Next pulls the next Trigram
func (t *Trigramizer) ReadTrigram() (Trigram, error) {
	// fill the last buffer

	for len(t.last) < t.n-1 {
		w, err := t.in.ReadWord()
		if err != nil {
			return nil, err
		}
		t.last = append(t.last, w)
	}

	c, err := t.in.ReadWord()
	if err != nil {
		return nil, err
	}

	// build the ngram

	out := []string{}
	out = append(out, t.last...)
	out = append(out, c)

	// remember the last n-1

	t.last = append(t.last[1:], c)

	return out, nil
}

// Trigrams is the database of Trigrams and generator of text
type Trigrams struct {
	n int
	// first two words + third words
	data map[string][]string
	lock sync.RWMutex
}

// NewTrigrams makes an empty Trigrams
func NewTrigrams(n int) *Trigrams {
	if n < 2 {
		panic("less than 3")
	}

	out := &Trigrams{
		n:    n,
		data: map[string][]string{},
	}
	return out
}

// InputTrigrams adds to the database
func (tg *Trigrams) InputTrigrams(input []Trigram) error {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	for _, ngram := range input {
		if len(ngram) != tg.n {
			panic("bad")
		}

		prefix := ngram.Prefix(tg.n - 1)
		key := strings.Join(prefix, " ")

		nthWords := tg.data[key]
		nthWords = append(nthWords, ngram.Last())

		tg.data[key] = nthWords
	}

	return nil
}

// GenerateN generates N words of text. If the start string contains 2 words, it will start from that.
func (tg *Trigrams) GenerateN(start string, length int) (string, error) {
	tg.lock.RLock()
	defer tg.lock.RUnlock()

	// first n-1 words are key to n
	key := ""
	if start != "" {
		key = start
	} else {
		for key = range tg.data {
			break
		}
	}

	if key == "" {
		return "", errors.New("no data")
	}

	out := key
	for i := length; i > 0; i-- {
		nthWords := tg.data[key]
		if nthWords == nil {
			break
		}
		rn := rand.Intn(len(nthWords))
		nextWord := nthWords[rn]
		out += " " + nextWord
		key = strings.SplitN(key, " ", 2)[1] + " " + nextWord
	}
	return out, nil
}

// LearnTextStream is a somewhat efficient way of parsing text and storing as Trigrams
func LearnTextStream(tg *Trigrams, size int, stream io.RuneReader, keep string) (int, error) {
	tz := &Trigramizer{n: size, in: &Wordizer{in: stream, keep: keep}}

	batch := make([]Trigram, 0, 100)

	n := 0
	for {
		t, err := tz.ReadTrigram()
		if err != nil {
			if err == io.EOF {
				// done
				break
			}
			// real error
			return n, nil
		}
		batch = append(batch, t)
		n++

		if n%100 == 0 {
			err := tg.InputTrigrams(batch)
			if err != nil {
				return n, err
			}
			batch = batch[0:0]
		}
	}

	if len(batch) > 0 {
		err := tg.InputTrigrams(batch)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

// LearnTextStream sends text to the Trigrams.
func LearnTextString(tg *Trigrams, size int, text string, keep string) (int, error) {
	stream := strings.NewReader(text)
	return LearnTextStream(tg, size, stream, keep)
}
