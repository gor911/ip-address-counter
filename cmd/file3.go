package cmd

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func Run3() {
	f, err := os.Open("resources/ip-addresses.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	const chunkSize = 40 // e.g. 4 KB per chunk
	var leftover []byte  // carry any partial line
	buf := make([]byte, chunkSize)

	for {
		n, err := f.Read(buf)

		if n > 0 {
			// 1) combine leftover from last read with new bytes
			data := append(leftover, buf[:n]...)

			// 2) find the last newline in data
			cut := bytes.LastIndexByte(data, '\n')

			if cut == -1 {
				// no newline found: buffer too small for one line?
				// keep all data as leftover and read more
				leftover = data
			} else {
				// full-chunk is data up to and including that newline
				full := data[:cut+1]
				processChunk(full)

				// leftover is any bytes after that newline
				leftover = data[cut+1:]
			}
		}

		if err != nil {
			if err == io.EOF {
				spew.Dump(len(leftover))
				// If the file does not end with a newline.
				if len(leftover) > 0 {
					processChunk(leftover)
				}

				break
			}

			log.Fatal(err)
		}
	}
}

// processChunk handles a chunk of complete lines.
// Here we just print how many lines and bytes we got.
func processChunk(chunk []byte) {
	lines := bytes.Count(chunk, []byte{'\n'})
	fmt.Printf("Got %d lines (%d bytes)\n", lines, len(chunk))
	//fmt.Println(string(chunk))

	ipAddresses := strings.Fields(string(chunk))

	for _, ipAddress := range ipAddresses {
		numbers := strings.Split(ipAddress, ".")

		firstNumber, _ := strconv.Atoi(numbers[0])
		secondNumber, _ := strconv.Atoi(numbers[1])
		thirdNumber, _ := strconv.Atoi(numbers[2])
		fourthNumber, _ := strconv.Atoi(numbers[3])

		a := 1<<24*firstNumber + 1<<16*secondNumber + 1<<8*thirdNumber + fourthNumber
		fmt.Println(ipAddress, a)
	}

	// You can convert to string: text := string(chunk)
	// and work with each line if you like.
}
