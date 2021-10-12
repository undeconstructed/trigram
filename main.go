package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

type Trigrams struct {
	data map[string][]string
	lock sync.RWMutex
}

func NewTrigrams() *Trigrams {
	out := &Trigrams{
		data: map[string][]string{},
	}
	return out
}

func (tg *Trigrams) LearnText(text string) (int, error) {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	count := 0
	var a, b, c string
	a, text = nextWord(text)
	b, text = nextWord(text)
	if a == "" || b == "" {
		return 0, nil
	}
	for {
		c, text = nextWord(text)
		if c == "" {
			break
		}
		start := a + " " + b
		list := tg.data[start]
		list = append(list, c)
		tg.data[start] = list
		count++
		a = b
		b = c
	}
	return count, nil
}

func nextWord(text string) (word string, rest string) {
	n, char := 0, '.'
	for n, char = range text {
		if char >= 'a' && char <= 'z' {
			word = word + string(char)
		} else if char >= 'A' && char <= 'Z' {
			// XXX - change case ?
			word = word + string(char)
		} else if char == '\'' {
			// XXX - strip this?
			word = word + string(char)
		} else {
			if len(word) > 0 {
				break
			}
			// else totally ignore this char
		}
	}
	if n < len(text) {
		rest = text[n+1:]
	} else {
		rest = ""
	}
	return word, rest
}

func (tg *Trigrams) GenerateN(length int) (string, error) {
	tg.lock.RLock()
	defer tg.lock.RUnlock()

	key := ""
	for key = range tg.data {
		break
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
		text, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		res, err := trigrams.LearnText(string(text))
		if err != nil {
			w.WriteHeader(400)
			return
		}

		fmt.Fprintf(w, "%v", res)
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(400)
			return
		}

		res, err := trigrams.GenerateN(100)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		fmt.Fprintf(w, "%v", res)
	})

	http.HandleFunc("/grams", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", trigrams)
	})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
