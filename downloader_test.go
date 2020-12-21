package main

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadConcurrent(t *testing.T) {

	// create `testdata` directory
	_ = os.Mkdir("testdata", os.ModeDir)

	// create the test file
	tmpfile, err := ioutil.TempFile("testdata", "test2mb")
	if err != nil {
		t.Fatal("creating temp file:", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	buff := make([]byte, 2*(1<<20))
	rand.Read(buff)
	if _, err := tmpfile.Write(buff); err != nil {
		t.Fatal("write to temp file:", err)
	}

	// spin up the server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, tmpfile.Name())
	}))
	defer ts.Close()

	c := ts.Client()

	resp, err := c.Head(ts.URL)

	if err != nil {
		t.Fatal("head request:", err)
	}

	down2mb := "testdata/down2mb"
	err = DownloadConcurrent(ts.URL, down2mb, 8, resp.ContentLength, &WriteCounter{})

	if err != nil {
		t.Fatal("downloading file:", err)
	}

	defer os.Remove(down2mb)

	f, err := ioutil.ReadFile(down2mb)
	if err != nil {
		t.Fatal("opening downloaded file:", err)
	}

	if !bytes.Equal(f, buff) {
		t.Error("file content and downloaded bits are not same")
	}
}
