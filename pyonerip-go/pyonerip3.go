// Sure, here is an example implementation that downloads and combines the .ts files in parallel using goroutines and buffers:

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

func downloadTsFile(url string, wg *sync.WaitGroup, index int, tsFilesCh chan<- *tsFile) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading file %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	tsFile := &tsFile{index: index}
	_, err = io.Copy(tsFile, resp.Body)
	if err != nil {
		fmt.Printf("Error copying response body: %v\n", err)
		return
	}
	tsFilesCh <- tsFile
}

func combineTsFiles(tsFiles []*tsFile, outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, tsFile := range tsFiles {
		_, err = f.Write(tsFile.buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadAndCombine(m3u8Url string, outFile string) error {
	tsFiles, err := parseM3u8(m3u8Url)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	tsFilesCh := make(chan *tsFile, len(tsFiles))

	for i, tsUrl := range tsFiles {
		wg.Add(1)
		go downloadTsFile(tsUrl, &wg, i, tsFilesCh)
	}

	wg.Wait()
	close(tsFilesCh)

	var tsFileArr []*tsFile
	for tsFile := range tsFilesCh {
		tsFileArr = append(tsFileArr, tsFile)
	}

	return combineTsFiles(tsFileArr, outFile)
}

func main() {
	err := downloadAndCombine("http://example.com/playlist.m3u8", "output.ts")
	if err != nil {
		fmt.Printf("Error downloading and combining: %v\n", err)
		return
	}
	fmt.Println("Download and combine successful")
}

type tsFile struct {
	buf   bytes.Buffer
	index int
}

func (t *tsFile) Write(p []byte) (int, error) {
	return t.buf.Write(p)
}

func (t *tsFile) Size() int64 {
	return int64(t.buf.Len())
}

func (t *tsFile) Name() string {
	return "file_" + strconv.Itoa(t.index) + ".ts"
}


// This implementation creates a channel to receive the downloaded .ts files, and uses a sync.WaitGroup to wait for all the downloads to complete before closing the channel. Then, the downloaded files are combined into a single file using buffers and written to disk. The downloadTsFile function is called in a goroutine to allow parallel downloads.
