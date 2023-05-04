package main

import (
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

// downloadFile downloads a single file from the given URL
// and saves it to disk using the file's name as provided by the URL.
// It signals the filename on the given channel when the download is complete.
func downloadFile(url string, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	filename := url[strings.LastIndex(url, "/")+1:]
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}
	ch <- filename
}

// mergeFiles combines the input files with the given filenames
// into a single output file with the provided filename.
// It sorts the input files by filename before merging them,
// and uses a buffer to optimize memory usage during the merge.
func mergeFiles(filenames []string, outputFilename string) error {
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	sort.Strings(filenames)
	buf := make([]byte, 32*1024)
	for _, filename := range filenames {
		inputFile, err := os.Open(filename)
		if err != nil {
			return err
		}
		_, err = io.CopyBuffer(outputFile, inputFile, buf)
		if err != nil {
			inputFile.Close()
			return err
		}
		inputFile.Close()
	}
	return nil
}

// main is the entry point for the program.
// It defines the list of file URLs to download,
// creates a wait group and a channel for the download goroutines,
// starts a download goroutine for each file URL,
// and waits for all downloads to complete.
// Once all downloads are complete, it sorts the downloaded files by filename
// and merges them into a single output file using a buffer.
func main() {
	fileUrls := []string{
		"http://example.com/file1.ts",
		"http://example.com/file2.ts",
		"http://example.com/file3.ts",
	}
	var wg sync.WaitGroup
	filenameCh := make(chan string, len(fileUrls))
	for _, fileUrl := range fileUrls {
		wg.Add(1)
		go downloadFile(fileUrl, &wg, filenameCh)
	}
	wg.Wait()
	close(filenameCh)
	filenames := []string{}
	for filename := range filenameCh {
		filenames = append(filenames, filename)
	}
	mergeFiles(filenames, "output.ts")
}
