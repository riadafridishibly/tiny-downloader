package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

// WriteCounter tracks the number of bytes downloaded
type WriteCounter struct {
	LastTime int64
	Total    uint64
}

func (wc *WriteCounter) Write(buf []byte) (int, error) {
	n := len(buf)
	atomic.AddUint64(&wc.Total, uint64(n))
	return n, nil
}

// GetCount returns the total bytes downloaded
func (wc *WriteCounter) GetCount() uint64 {
	return atomic.LoadUint64(&wc.Total)
}

// GetSpeed calculates the speed of download
func (wc *WriteCounter) GetSpeed() string {
	now := time.Now().UnixNano()
	seconds := float64(now-wc.LastTime) * 1e-9
	bytesWritten := wc.GetCount()
	return humanize.Bytes(uint64(float64(bytesWritten)/seconds)) + "/s"
}

// ShowProgress prints the download progress concurrently
func ShowProgress(wc *WriteCounter, contentLength uint64, filename string, done chan struct{}) {
	for {
		written := wc.GetCount()
		fmt.Printf("\r[%s] Downloaded: [%s/%s] (%.2f%%) DL: %s",
			filename,
			humanize.Bytes(written),
			humanize.Bytes(contentLength),
			float64(written)/float64(contentLength)*100,
			wc.GetSpeed(),
		)
		if written >= contentLength {
			fmt.Println()
			done <- struct{}{}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
