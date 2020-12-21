package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Download downloads a single file
func Download(url, filename string, startAt, count int64, writeCounter *WriteCounter) error {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	rangeValue := fmt.Sprintf("bytes=%d-%d", startAt, startAt+count-1)
	req.Header.Set("Range", rangeValue)

	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outfile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outfile.Close()

	src := io.TeeReader(resp.Body, writeCounter)

	_, err = io.Copy(outfile, src)
	if err != nil {
		return err
	}

	return nil
}

// DownloadConcurrent downloads file using n goroutine
func DownloadConcurrent(url, filename string, n int, contentLength int64, wc *WriteCounter) error {
	wg := sync.WaitGroup{}

	dl := func(url, filename string, start, count int64) {
		defer wg.Done()
		finfo, err := os.Stat(filename)
		if err == nil { // partial file found
			if finfo.Size() < count { // full file yet to be downloaded
				start += finfo.Size()
				count -= finfo.Size()
			} else {
				return
			}
			wc.Add(finfo.Size())
		}
		err = Download(url, filename, start, count, wc)
		if err != nil {
			log.Fatal(err)
		}
	}

	if n <= 0 || n > 16 {
		n = 8
	}

	aggregateNeeded := false

	if n > 1 {
		aggregateNeeded = true
		size := int64(math.Ceil(float64(contentLength) / float64(n)))
		for i := 0; i < n; i++ {
			wg.Add(1)
			go dl(url, filename+".part"+strconv.Itoa(i), int64(i)*size, size)
		}
	} else {
		wg.Add(1)
		go dl(url, filename, 0, contentLength)
	}

	wg.Wait() // wait for all downloader to finish

	if aggregateNeeded {
		err := AggregateFiles(filename, n)

		if err != nil {
			return err
		}
	}
	return nil
}

// AggregateFiles aggregate the part files to a single file
func AggregateFiles(filename string, n int) error {
	mainFile, err := os.OpenFile(filename+".part0", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer mainFile.Close()

	for i := 1; i < n; i++ {
		partFilename := filename + ".part" + strconv.Itoa(i)
		part, err := os.Open(partFilename)
		if err != nil {
			return err
		}
		io.Copy(mainFile, part)
		part.Close()

		os.Remove(partFilename)
	}

	err = os.Rename(filename+".part0", filename)
	if err != nil {
		return err
	}

	return nil
}
