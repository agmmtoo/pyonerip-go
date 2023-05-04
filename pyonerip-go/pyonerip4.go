package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
)

func downloadFile(url string, wg *sync.WaitGroup, ch chan<- string) {
	// Decrement the WaitGroup counter when the function completes
	defer wg.Done()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %v\n", url, err)
		return
	}

	// Send the HTTP request and get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request to %s: %v\n", url, err)
		return
	}

	// Ensure the response body is closed when the function completes
	defer resp.Body.Close()

	// Get the filename from the URL
	filename := url[strings.LastIndex(url, "/")+1:]

	// Create a new file with the given filename
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}

	// Ensure the file is closed when the function completes
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", url, err)
		return
	}

	// Send the filename on the channel to signal that the download is complete
	ch <- filename
}

func mergeFiles(filenames []string, outputFilename string) error {
	// Create the output file
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("Error creating output file %s: %v", outputFilename, err)
	}
	defer outputFile.Close()

	// Sort the filenames
	sort.Strings(filenames)

	// Create a buffer for reading the input files
	buf := make([]byte, 32*1024)

	// Loop over the input files and copy their contents to the output file
	for _, filename := range filenames {
		// Open the input file
		inputFile, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("Error opening input file %s: %v", filename, err)
		}

		// Copy the contents of the input file to the output file
		_, err = io.CopyBuffer(outputFile, inputFile, buf)
		if err != nil {
			inputFile.Close()
			return fmt.Errorf("Error copying from %s to %s: %v", filename, outputFilename, err)
		}

		// Close the input file
		inputFile.Close()
	}

	return nil
}

func main() {
	// Define the list of file URLs to download
	fileUrls := []string{
		"http://example.com/file1.ts",
		"http://example.com/file2.ts",
		"http://example.com/file3.ts",
	}

	// Create a wait group to ensure all downloads are complete
	var wg sync.WaitGroup

	// Create a channel to receive filenames from the download goroutines
	filenameCh := make(chan string, len(fileUrls))

	// Loop over the file URLs and start a download goroutine for each one
	for _, fileUrl := range fileUrls {
		// Add a download to the wait group
		wg.Add(1)

		// Start a download goroutine
		go downloadFile(fileUrl, &wg, filenameCh)

		// Close the filename channel
		close(filenameCh)

		// Collect the filenames from the channel
		var filenames []string
		for filename := range filenameCh {
			filenames = append(filenames, filename)
		}

		// Define the output filename
		outputFilename := "output.ts"

		// Merge the downloaded files into the output file
		err := mergeFiles(filenames, outputFilename)
		if err != nil {
			fmt.Printf("Error merging files: %v\n", err)
			return
		}

		fmt.Printf("Merged files into %s\n", outputFilename)
	}
}

// In the above program, we first define the list of file URLs to download. Then, we create a wait group to ensure all downloads are complete, and a channel to receive filenames from the download goroutines.

// We loop over the file URLs and start a download goroutine for each one. Each download goroutine uses the downloadFile function to download the file, and sends the filename on the channel when the download is complete.

// After all downloads are complete, we close the filename channel and collect the filenames from the channel.

// Then, we define the output filename and call the mergeFiles function to merge the downloaded files into the output file. The mergeFiles function sorts the filenames and uses a buffer to read and write the files in chunks, to conserve memory.

// Finally, we print a message indicating that the files have been merged into the output file.
