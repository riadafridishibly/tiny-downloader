package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Download a file
func Download(url, filename string, startAt, count uint64) error {
	start := strconv.FormatUint(startAt, 10)
	end := strconv.FormatUint(startAt+count-1, 10)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Ranges", "bytes="+start+"-"+end)

	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

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

	fmt.Printf("%v\n", filename)
	fmt.Println("Ranges: ", isRangeSupported, "\nContent-Length: ", contentLength)

	if filename == "" {
		filename = "output"
	}

	err = Download(url, filename, 0, uint64(contentLength))

	if err != nil {
		panic(err)
	}
}
