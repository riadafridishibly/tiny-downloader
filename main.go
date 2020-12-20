package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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

	resp, err := http.Head(url)

	if err != nil { // failed HEAD request, try GET request
		resp, err = http.Get(url)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	// check if server supports byte ranges
	isRangeSupported := false

	if val, ok := resp.Header["Accept-Ranges"]; ok {
		if val[0] == "bytes" { // server supports ranges
			isRangeSupported = true
		}
	}

	contentLength := resp.ContentLength
	filename := ""

	if val, ok := resp.Header["Content-Disposition"]; ok {
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
		filename = strings.ReplaceAll(filename, "%20", " ")
		if filename == "/" || filename == "." {
			filename = "output"
		}
	}

	writeCounter := &WriteCounter{Total: 0, LastTime: time.Now().UnixNano()}

	done := make(chan struct{})

	go ShowProgress(writeCounter, uint64(contentLength), filename, done)

	if isRangeSupported && contentLength > 1024 {
		DownloadConcurrent(url, filename, ngoroutine, contentLength, writeCounter)
	} else {
		DownloadConcurrent(url, filename, 1, contentLength, writeCounter)
	}
	// wait for progress bar to finish
	<-done
}
