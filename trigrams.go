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
// XXX - is it weird to use an array?
type Trigram [3]string

func (t Trigram) String() string {
	ss := [3]string(t)
	return strings.Join(ss[:], " ")
}

// TrigramReader has ReadTrigram
type TrigramReader interface {
	TrigramWord() (Trigram, error)
}

// Trigramizer pulls successive Trigrams from a stream
type Trigramizer struct {
	a, b string
	in   WordReader
}

// Next pulls the next Trigram
func (t *Trigramizer) ReadTrigram() (out Trigram, err error) {
	a, b, c := t.a, t.b, ""
	if a == "" {
		// initialisation
		a, err = t.in.ReadWord()
		if err != nil {
			return
		}
		t.a = a
		b, err = t.in.ReadWord()
		if err != nil {
			return
		}
		t.b = b
	}
	if b == "" {
		// no more data
		return Trigram{}, io.EOF
	}

	c, err = t.in.ReadWord()
	if err != nil {
		return
	}

	out = Trigram{a, b, c}
	t.a = b
	t.b = c

	return out, nil
}

// Trigrams is the database of Trigrams and generator of text
type Trigrams struct {
	// first two words + third words
	data map[string][]string
	lock sync.RWMutex
}

// NewTrigrams makes an empty Trigrams
func NewTrigrams() *Trigrams {
	out := &Trigrams{
		data: map[string][]string{},
	}
	return out
}

// InputTrigrams adds to the database
func (tg *Trigrams) InputTrigrams(input []Trigram) error {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	for _, i := range input {
		key := i[0] + " " + i[1]
		list := tg.data[key]
		list = append(list, i[2])
		tg.data[key] = list
	}

	return nil
}

// LearnTextStream is a somewhat efficient way of parsing text and storing as Trigrams
func (tg *Trigrams) LearnTextStream(stream io.RuneReader, keep string) (int, error) {
	tz := &Trigramizer{in: &Wordizer{in: stream, keep: keep}}

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

// LearnTextStream is nothing.
func (tg *Trigrams) LearnTextString(text string, keep string) (int, error) {
	stream := strings.NewReader(text)
	return tg.LearnTextStream(stream, keep)
}

// GenerateN generates N words of text. If the start string contains 2 words, it will start from that.
func (tg *Trigrams) GenerateN(start string, length int) (string, error) {
	tg.lock.RLock()
	defer tg.lock.RUnlock()

	// first two words are key to third
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
		list := tg.data[key]
		if list == nil {
			break
		}
		rn := rand.Intn(len(list))
		nextWord := list[rn]
		out += " " + nextWord
		key = strings.Split(key, " ")[1] + " " + nextWord
	}
	return out, nil
}
