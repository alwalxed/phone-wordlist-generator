package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	bufferSize = 32 * 1024 * 1024
	numWorkers = 4
	batchSize  = 100000
)

var (
	outputDir = flag.String("output", "", "Output directory (default: ~/.wordlist-generator)")
	help      = flag.Bool("help", false, "Show help message")
)

type batch struct {
	start int
	end   int
}

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *outputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			*outputDir = "./wordlists"
		} else {
			*outputDir = filepath.Join(homeDir, ".wordlist-generator")
		}
	}

	fmt.Print("Enter prefix: ")

	reader := bufio.NewReader(os.Stdin)
	prefix, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading input:", err)
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		log.Fatal("Prefix cannot be empty")
	}

	if len(prefix) > 50 {
		log.Fatal("Prefix too long (max 50 characters)")
	}

	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(prefix, char) {
			log.Fatal("Prefix contains invalid character:", char)
		}
	}

	start := time.Now()
	if err := generateWordlistUltraFast(prefix); err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Printf("Completed in: %v\n", time.Since(start))
}

func showHelp() {
	fmt.Printf("Wordlist Generator\n\n")
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Printf("Options:\n")
	flag.PrintDefaults()
	fmt.Printf("\nExample:\n")
	fmt.Printf("  %s -output ./my-wordlists\n", os.Args[0])
	fmt.Printf("\nDefault output directory: ~/.wordlist-generator\n")
}

func generateWordlistUltraFast(prefix string) error {
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filename := fmt.Sprintf("%s-XXX-XXXX.txt", prefix)
	filepath := filepath.Join(*outputDir, filename)

	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("file '%s' already exists", filename)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file properly: %v\n", closeErr)
		}
	}()

	writer := bufio.NewWriterSize(file, bufferSize)
	defer func() {
		if flushErr := writer.Flush(); flushErr != nil {
			fmt.Printf("Warning: failed to flush buffer: %v\n", flushErr)
		}
	}()

	batchChan := make(chan batch, numWorkers*2)
	resultChan := make(chan []byte, numWorkers*2)
	errorChan := make(chan error, numWorkers)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			workerBuf := make([]byte, 0, batchSize*(len(prefix)+8))

			for b := range batchChan {
				workerBuf = workerBuf[:0]

				if err := generateBatch(prefix, b.start, b.end, &workerBuf); err != nil {
					select {
					case errorChan <- fmt.Errorf("worker %d error: %w", workerID, err):
					default:
					}
					return
				}

				batchData := make([]byte, len(workerBuf))
				copy(batchData, workerBuf)

				select {
				case resultChan <- batchData:
				case <-time.After(10 * time.Second):
					select {
					case errorChan <- fmt.Errorf("worker %d: timeout sending result", workerID):
					default:
					}
					return
				}
			}
		}(i)
	}

	writerDone := make(chan error, 1)
	totalWritten := 0

	go func() {
		defer close(writerDone)
		for result := range resultChan {
			if _, err := writer.Write(result); err != nil {
				writerDone <- fmt.Errorf("failed to write batch: %w", err)
				return
			}
			totalWritten += countLines(result)
		}
		writerDone <- nil
	}()

	go func() {
		defer close(batchChan)
		for i := 0; i < 10000000; i += batchSize {
			end := i + batchSize
			if end > 10000000 {
				end = 10000000
			}

			select {
			case batchChan <- batch{start: i, end: end}:
			case <-time.After(5 * time.Second):
				select {
				case errorChan <- fmt.Errorf("timeout sending batch %d-%d", i, end):
				default:
				}
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case err := <-errorChan:
		return fmt.Errorf("generation failed: %w", err)
	case err := <-writerDone:
		if err != nil {
			return err
		}
	}

	if totalWritten != 10000000 {
		return fmt.Errorf("safety check failed: expected 10000000 lines, wrote %d", totalWritten)
	}

	fmt.Printf("Generated %d combinations in %s\n", totalWritten, filepath)
	return nil
}

func generateBatch(prefix string, start, end int, buf *[]byte) error {
	if start < 0 || end < start || end > 10000000 {
		return fmt.Errorf("invalid batch range: %d-%d", start, end)
	}

	prefixBytes := []byte(prefix)

	for i := start; i < end; i++ {
		if len(*buf)+len(prefix)+8 > cap(*buf) {
			return fmt.Errorf("buffer overflow prevented at line %d", i)
		}

		*buf = append(*buf, prefixBytes...)

		num := i
		if num < 0 || num >= 10000000 {
			return fmt.Errorf("number out of range: %d", num)
		}

		digitStart := len(*buf)
		*buf = append(*buf, "0000000"...)

		for j := 6; j >= 0; j-- {
			(*buf)[digitStart+j] = byte('0' + num%10)
			num /= 10
		}

		*buf = append(*buf, '\n')
	}

	return nil
}

func countLines(data []byte) int {
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}
