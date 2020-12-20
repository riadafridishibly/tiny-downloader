package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Usage prints the usages of this application
var Usage = func() {
	h := ""
	h += fmt.Sprintf("Usages:")
	h += fmt.Sprintf("\t%s [-n N] <url>\n", os.Args[0])
	h += fmt.Sprintf("\t%-10s%-10s%s", "-n", "<n>", "Number of goroutine for downloading\n")
	fmt.Fprintf(os.Stderr, h)
}

func main() {
	flag.Usage = Usage
	var ngoroutine int
	flag.IntVar(&ngoroutine, "n", 8, "Number of goroutine for downloading")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	url := flag.Arg(0)

	header, err := http.Head(url)
	if err != nil {
		header, err = http.Get(url)
		header.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	// check if server supports byte ranges

	isRangeSupported := false

	if val, ok := header.Header["Accept-Ranges"]; ok {
		if val[0] == "bytes" { // server supports ranges
			isRangeSupported = true
		}
	}

	contentLength := header.ContentLength

	filename := ""

	if val, ok := header.Header["Content-Disposition"]; ok {
		if idx := strings.Index(val[0], "filename"); idx != -1 {
			idx += len("filename=\"")
			i := idx
			for i < len(val[0]) && val[0][i] != '"' {
				i++
			}
			filename = val[0][idx:i]
		}
	}

	if filename == "" {
		filename = path.Base(url)
		if filename == "/" || filename == "." {
			filename = "output"
		}
	}

	writeCounter := &WriteCounter{Total: 0, LastTime: time.Now().UnixNano()}

	done := make(chan struct{})

	go ShowProgress(writeCounter, uint64(contentLength), filename, done)

	// err = Download(url, filename, 0, uint64(contentLength), writeCounter)

	wg := sync.WaitGroup{}

	dl := func(url, filename string, start, count uint64) {
		defer wg.Done()
		err := Download(url, filename, start, count, writeCounter)
		if err != nil {
			log.Fatal(err)
		}
	}

	n := ngoroutine

	if n < 0 || n > 16 {
		n = 8
	}

	aggregateNeeded := false

	if isRangeSupported && contentLength > 1024 && n > 1 {
		aggregateNeeded = true
		size := uint64(math.Ceil(float64(contentLength) / float64(n)))
		for i := 0; i < n; i++ {
			wg.Add(1)
			go dl(url, filename+".part"+strconv.Itoa(i), uint64(i)*size, size)
		}
	} else {
		wg.Add(1)
		go dl(url, filename, 0, uint64(contentLength))
	}

	wg.Wait()

	if aggregateNeeded {
		err := AggregateFiles(filename, n)

		if err != nil {
			log.Fatal(err)
		}
	}
	// wait for progress bar to finish
	<-done
}
