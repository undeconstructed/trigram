# trigram

```
$ go run .
$ for f in *.txt ; do curl --data-binary @$f -H"Content-Type: text/plain" localhost:8080/learn ; done
$ curl -v localhost:8080/generate
```
