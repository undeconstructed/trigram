# trigram

## Running

Typical:
```
$ go run . -keep="'"
$ for f in *.txt ; do curl --data-binary @$f -H"Content-Type: text/plain" localhost:8080/learn ; done
$ curl -s 'localhost:8080/generate'
```

Command options:
```
$ go run . -keep=",.?\!:;'\"“”’" -addr="0.0.0.0:8080"
```

Generate from a starting point:
```
$ curl -s 'localhost:8080/grams?start=to%20my'
```

## Function

Text streams are broken into words, then into trigrams, then batch-posted into the database. The parsing is done in the HTTP routines, so the database is only locked when the trigrams are being inserted.

Generation happens in a read lock.

## Data structure

Just `map[string][]string` - ```"first second" -> [ "third", "third", "third" ]```

Combining the first two makes lookup very quick, but means duplicating first words. Also means it's not easy to start generation from a given word, but that wasn't mentioned in the spec anyway.

As it happens, the flatter structure would make serialisation easier, and the combination of the first two words would make sharding more effective ..

Third words can be stored multiple times in the same list (e.g. "on the" => ["beach", "floor", "beach"]). Counting instances would potentially be more compact, but would make generation harder. On a quick check the average was about 2.5 words in each list, so it's not immediately a problem.

## Issues

* Frontend is very basic, it just checks the bare minimum of things, and errors are not communicated.
* Not sure about punctuation. The examples show some, but also suggest stripping it out? I made an option to control what is kept, in addition to letters.
* The spec says text/plain input, but shows curl with no content type set. I explicitly checked for text/plain.
* LearnTextStream is not testable - the batching should be separated out from the reading/writing.
* Global random source is used, so again not testable.
