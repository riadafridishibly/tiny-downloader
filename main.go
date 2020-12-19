package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

type WriteCounter struct {
	LastTime  int64
	LastWrite uint64
	Total     uint64
}

func (wc *WriteCounter) Write(buf []byte) (int, error) {
	n := len(buf)
	atomic.AddUint64(&wc.Total, uint64(n))
	return n, nil
}

func (wc *WriteCounter) GetCount() uint64 {
	return atomic.LoadUint64(&wc.Total)
}

func (wc *WriteCounter) GetSpeed() string {
	now := time.Now().UnixNano()
	seconds := float64(now-wc.LastTime) * 1e-9
	bytesWritten := wc.GetCount() - wc.LastWrite

	// wc.LastWrite = wc.GetCount()
	// wc.LastTime = time.Now().UnixNano()

	return humanize.Bytes(uint64(float64(bytesWritten)/seconds)) + "/s"
}

// Download a file
func Download(url, filename string, startAt, count uint64, writeCounter *WriteCounter) error {
	start := strconv.FormatUint(startAt, 10)
	end := strconv.FormatUint(startAt+count-1, 10)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Range", "bytes="+start+"-"+end)

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

	src := io.TeeReader(resp.Body, writeCounter)

	_, err = io.Copy(outfile, src)
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

	if filename == "" {
		filename = "output"
	}

	// TODO: spawn some goroutine
	if isRangeSupported {
	}

	writeCounter := &WriteCounter{Total: 0, LastTime: time.Now().UnixNano()}

	done := make(chan struct{})
	// show a progress bar
	// TODO: shift this to a standalone function
	go func(wc *WriteCounter, contentLength uint64) {
		for {
			written := wc.GetCount()
			fmt.Printf("\r[%s] Downloaded: %.2f%% %s", filename, (float64(written)/float64(contentLength))*100, wc.GetSpeed())
			if written >= contentLength {
				fmt.Println()
				done <- struct{}{}
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(writeCounter, uint64(contentLength))

	// err = Download(url, filename, 0, uint64(contentLength), writeCounter)

	wg := sync.WaitGroup{}

	dl := func(url, filename string, start, count uint64) {
		defer wg.Done()
		err := Download(url, filename, start, count, writeCounter)
		if err != nil {
			panic(err)
		}
	}

	n := 8

	if isRangeSupported {
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
	// wait for progress bar to finish
	<-done
}
