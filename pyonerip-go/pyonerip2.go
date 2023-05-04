// You are correct, using io.ReadAll to read the entire file into memory at once may not be the most efficient solution, especially for very large files. A better approach would be to use a streaming approach, where you read the input file in chunks and write those chunks to the output file as you go.

// One way to do this in Go is to use the bufio package to read and write the files in chunks. Here's an updated version of the code that uses bufio to stream the input and output files:

package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

const (
	m3u8URL        = "https://example.com/stream.m3u8"
	tsFilePrefix   = "https://example.com/"
	outputFileName = "output.ts"
	maxConcurrency = 10
)

func main() {
	// Download the m3u8 file.
	resp, err := http.Get(m3u8URL)
	if err != nil {
		fmt.Printf("Failed to download m3u8 file: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Create the output file.
	outfile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		return
	}
	defer outfile.Close()

	// Create a scanner to read the m3u8 file line by line.
	scanner := bufio.NewScanner(resp.Body)

	// Create a channel to receive the URLs of the .ts files.
	urls := make(chan string)

	// Create a wait group to track when all the downloads are done.
	var wg sync.WaitGroup

	// Start a fixed number of goroutines to download the .ts files in parallel.
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			for url := range urls {
				// Download the .ts file.
				resp, err := http.Get(url)
				if err != nil {
					fmt.Printf("Failed to download %s: %v\n", url, err)
					continue
				}
				defer resp.Body.Close()

				// Copy the .ts file to the output file.
				_, err = io.Copy(outfile, resp.Body)
				if err != nil {
					fmt.Printf("Failed to write %s to output file: %v\n", url, err)
					continue
				}

				// Signal that the download is done.
				wg.Done()
			}
		}()
	}

	// Read the m3u8 file line by line.
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line is a .ts file URL.
		if strings.HasPrefix(line, tsFilePrefix) {
			// Add the .ts file URL to the URLs channel.
			urls <- line
			// Increment the wait group to track the new download.
			wg.Add(1)
		}
	}

	// Close the URLs channel to signal that all the URLs have been processed.
	close(urls)

	// Wait for all the downloads to finish.
	wg.Wait()

	fmt.Println("Done.")
}

// In this version of the code, we use a bufio.Scanner to read the m3u8 file line by line, and check each line to see if it is a .ts file URL. If it is, we add the URL to a channel, and increment a wait group to track the new download.

// We then start a fixed number of goroutines to download the .ts files from the URLs in the channel. Each goroutine reads the response
