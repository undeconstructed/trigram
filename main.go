package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// const KeepRunes = ".,!\"':;"
const KeepRunes = "'.,"

// Trigram is just three strings
// XXX - is it weird to use an array?
type Trigram [3]string

// Trigramizer pulls successive Trigrams from a stream
type Trigramizer struct {
	a, b string
	in   io.RuneReader
}

// Next pulls the next Trigram
func (t *Trigramizer) Next() (Trigram, error) {
	if t.a == "" {
		// never rune
		t.a = nextWord(t.in)
		t.b = nextWord(t.in)
	}
	if t.b == "" {
		// no more data
		return Trigram{}, io.EOF
	}

	c := nextWord(t.in)
	if c == "" {
		// now that's the end
		t.b = c
		return Trigram{}, io.EOF
	}

	out := Trigram{t.a, t.b, c}
	t.a = t.b
	t.b = c

	return out, nil
}

// nextWord acquires a word from a stream
func nextWord(in io.RuneReader) string {
	word := ""
	for {
		rune, _, err := in.ReadRune()
		if err != nil {
			if err == io.EOF {
				return word
			}
			// TODO - better
			return ""
		}
		if rune >= 'a' && rune <= 'z' {
			word = word + string(rune)
		} else if rune >= 'A' && rune <= 'Z' {
			// XXX - change case ?
			word = word + string(rune)
		} else if strings.Contains(KeepRunes, string(rune)) {
			word = word + string(rune)
		} else {
			if len(word) > 0 {
				// found a word, so consider this runeacter a break
				break
			}
			// else totally ignore this rune
		}
	}
	return word
}

// Trigrams is the database of Trigrams and generator of text
type Trigrams struct {
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
func (tg *Trigrams) LearnTextStream(stream io.RuneReader) (int, error) {
	tz := &Trigramizer{in: stream}

	batch := make([]Trigram, 0, 100)

	n := 0
	for {
		t, err := tz.Next()
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
func (tg *Trigrams) LearnTextString(text string) (int, error) {
	stream := strings.NewReader(text)
	return tg.LearnTextStream(stream)
}

// GenerateN generates N words of text. If the start string contains 2 words, it will start from that.
func (tg *Trigrams) GenerateN(start string, length int) (string, error) {
	tg.lock.RLock()
	defer tg.lock.RUnlock()

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

func main() {
	addr := "0.0.0.0:8080"
	fmt.Printf("trigram: http://%s/\n", addr)

	// first two words + third words
	trigrams := NewTrigrams()

	http.HandleFunc("/learn", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(400)
			return
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(400)
			return
		}

		stream := bufio.NewReader(r.Body)
		n, err := trigrams.LearnTextStream(stream)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		fmt.Fprintf(w, "%v\n", n)
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(400)
			return
		}

		start := r.URL.Query().Get("start")
		lengthS := r.URL.Query().Get("length")
		length, err := strconv.Atoi(lengthS)
		if length < 2 {
			length = 100
		}

		res, err := trigrams.GenerateN(start, length)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		fmt.Fprintf(w, "%v\n", res)
	})

	http.HandleFunc("/grams", func(w http.ResponseWriter, r *http.Request) {
		// XXX - DEMONSTRATION PURPOSES - not locked
		fmt.Fprintf(w, "%v\n", trigrams.data)
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		// XXX - DEMONSTRATION PURPOSES - not locked
		prefixes := len(trigrams.data)
		total := 0
		for _, v := range trigrams.data {
			total += len(v)
		}
		fmt.Fprintf(w, "prefixes: %d, endings: %d, ratio: %f\n", prefixes, total, float64(total)/float64(prefixes))
	})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
