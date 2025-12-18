package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// -----------------------------
// Entry point
// -----------------------------

func main() {
	rand.Seed(time.Now().UnixNano())

	r := flag.Int("r", -1, "generate N random integers (N >= 10)")
	i := flag.String("i", "", "input file with integers")
	d := flag.String("d", "", "directory with input .txt files")
	flag.Parse()

	switch {
	case *r != -1:
		if err := runRandom(*r); err != nil {
			log.Fatal(err)
		}
	case *i != "":
		if err := runInputFile(*i); err != nil {
			log.Fatal(err)
		}
	case *d != "":
		if err := runDirectory(*d); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Usage: gosort -r N | -i file.txt | -d directory")
	}
}

// -----------------------------
// -r mode
// -----------------------------

func runRandom(n int) error {
	if n < 10 {
		return errors.New("N must be >= 10")
	}

	numbers := generateRandomNumbers(n)

	fmt.Println("Original numbers:")
	fmt.Println(numbers)

	processAndPrint(numbers)
	return nil
}

// -----------------------------
// -i mode
// -----------------------------

func runInputFile(path string) error {
	numbers, err := readNumbersFromFile(path)
	if err != nil {
		return err
	}

	if len(numbers) < 10 {
		return errors.New("input file must contain at least 10 valid integers")
	}

	fmt.Println("Original numbers:")
	fmt.Println(numbers)

	processAndPrint(numbers)
	return nil
}

// -----------------------------
// -d mode
// -----------------------------

func runDirectory(dir string) error {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return errors.New("invalid directory")
	}

	outputDir := dir + "_sorted_sandra_jaimy_241ADB123"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".txt" {
			continue
		}

		inputPath := filepath.Join(dir, f.Name())
		numbers, err := readNumbersFromFile(inputPath)
		if err != nil {
			return err
		}
		if len(numbers) < 10 {
			return fmt.Errorf("%s has fewer than 10 numbers", f.Name())
		}

		chunks := splitIntoChunks(numbers)
		sortedChunks := sortChunksConcurrently(chunks)
		result := mergeSortedChunks(sortedChunks)

		outputPath := filepath.Join(outputDir, f.Name())
		if err := writeNumbersToFile(outputPath, result); err != nil {
			return err
		}
	}

	return nil
}

// -----------------------------
// Shared processing
// -----------------------------

func processAndPrint(numbers []int) {
	chunks := splitIntoChunks(numbers)

	fmt.Println("\nChunks before sorting:")
	printChunks(chunks)

	sortedChunks := sortChunksConcurrently(chunks)

	fmt.Println("\nChunks after sorting:")
	printChunks(sortedChunks)

	result := mergeSortedChunks(sortedChunks)

	fmt.Println("\nFinal sorted result:")
	fmt.Println(result)
}

// -----------------------------
// Chunking logic
// -----------------------------

func splitIntoChunks(numbers []int) [][]int {
	n := len(numbers)

	numChunks := int(math.Ceil(math.Sqrt(float64(n))))
	if numChunks < 4 {
		numChunks = 4
	}

	chunks := make([][]int, numChunks)
	baseSize := n / numChunks
	remainder := n % numChunks

	index := 0
	for i := 0; i < numChunks; i++ {
		size := baseSize
		if i < remainder {
			size++
		}
		chunks[i] = numbers[index : index+size]
		index += size
	}
	return chunks
}

// -----------------------------
// Concurrent sorting
// -----------------------------

func sortChunksConcurrently(chunks [][]int) [][]int {
	var wg sync.WaitGroup
	wg.Add(len(chunks))

	for i := range chunks {
		go func(i int) {
			defer wg.Done()
			sort.Ints(chunks[i])
		}(i)
	}

	wg.Wait()
	return chunks
}

// -----------------------------
// Merge logic
// -----------------------------

func mergeSortedChunks(chunks [][]int) []int {
	result := []int{}

	indices := make([]int, len(chunks))

	for {
		minVal := 0
		minChunk := -1

		for i := range chunks {
			if indices[i] < len(chunks[i]) {
				val := chunks[i][indices[i]]
				if minChunk == -1 || val < minVal {
					minVal = val
					minChunk = i
				}
			}
		}

		if minChunk == -1 {
			break
		}

		result = append(result, minVal)
		indices[minChunk]++
	}

	return result
}

// -----------------------------
// Helpers
// -----------------------------

func generateRandomNumbers(n int) []int {
	nums := make([]int, n)
	for i := range nums {
		nums[i] = rand.Intn(1000) // range: 0–999
	}
	return nums
}

func readNumbersFromFile(path string) ([]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var numbers []int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		val, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %s", line)
		}
		numbers = append(numbers, val)
	}

	return numbers, scanner.Err()
}

func writeNumbersToFile(path string, nums []int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, n := range nums {
		fmt.Fprintln(file, n)
	}
	return nil
}

func printChunks(chunks [][]int) {
	for i, c := range chunks {
		fmt.Printf("Chunk %d: %v\n", i, c)
	}
}
