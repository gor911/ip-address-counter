package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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

	const chunkSize = 25 // e.g. 4 KB per chunk
	leftover := []byte{} // carry any partial line
	buf := make([]byte, chunkSize)

	for {
		n, err := f.Read(buf)

		if n > 0 {
			// 1) combine leftover from last read with new bytes
			data := append(leftover, buf[:n]...)

			// 2) find the last newline in data
			cut := bytes.LastIndexByte(data, '\n')

			if cut != -1 {
				// full-chunk is data up to and including that newline
				full := data[:cut+1]
				processChunk(full)

				// leftover is any bytes after that newline
				leftover = data[cut+1:]
			} else {
				// no newline found: buffer too small for one line?
				// keep all data as leftover and read more
				leftover = data
			}
		}

		if err != nil {
			if err == io.EOF {
				// process any remaining lines in leftover
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
	fmt.Println(string(chunk))
	// You can convert to string: text := string(chunk)
	// and work with each line if you like.
}
