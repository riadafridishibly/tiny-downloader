package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
)

// Download downloads a single file
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
