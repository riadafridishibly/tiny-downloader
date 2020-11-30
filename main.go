package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s <url>\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]

	header, err := http.Head(url)
	if err != nil {
		panic(err)
	}

	// check if supports Accept-Ranges

	isRangeSupported := false

	if val, ok := header.Header["Accept-Ranges"]; ok {
		if val[0] == "bytes" { // server supports ranges
			isRangeSupported = true
		}
	}

	contentLength := header.ContentLength

	fmt.Println("Ranges: ", isRangeSupported, "\nContent-Length: ", contentLength)
}
