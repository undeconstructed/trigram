package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	np := flag.Int("n", 3, "size of the ngrams")
	addrp := flag.String("addr", "0.0.0.0:8080", "bind addr")
	keepp := flag.String("keep", "',.", "to keep")
	flag.Parse()

	fmt.Printf("trigram (%s): http://%s/\n", *keepp, *addrp)

	trigrams := NewTrigrams(*np)

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
		n, err := LearnTextStream(trigrams, *np, stream, *keepp)
		if err != nil {
			w.WriteHeader(500)
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
		length, _ := strconv.Atoi(lengthS)
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
		start := r.URL.Query().Get("start")
		if start != "" {
			list := trigrams.data[start]
			fmt.Fprintf(w, "%v\n", list)
			return
		}

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

	err := http.ListenAndServe(*addrp, nil)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
