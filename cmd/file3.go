package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/bits"
	"os"
	"runtime"
	"sync"
)

func Run() {
	f, err := os.Open("resources/ip_addresses")

	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	const chunkSize = 100 * 1024 * 1024 // 100 MiB

	fullData := make([]byte, 0, chunkSize+16)
	buf := make([]byte, chunkSize)
	bitsArr := make([]uint64, wordsNeeded)

	for {
		n, err := f.Read(buf)

		if n > 0 {
			fullData = append(fullData, buf[:n]...)

			if cut := bytes.LastIndexByte(fullData, '\n'); cut >= 0 {
				processChunk(fullData[:cut], bitsArr, &mutex, &wg)

				copy(fullData, fullData[cut+1:])
				fullData = fullData[:len(fullData)-cut-1]
			}
		}

		if err != nil {
			if err == io.EOF {
				fmt.Println("fullData", len(fullData))
				// If the file does not end with a newline.
				if len(fullData) > 0 {
					fmt.Println("fullData", string(fullData))
					//wg.Add(1)
					processChunk(fullData, bitsArr, &mutex, &wg)
				}

				break
			}

			log.Fatal(err)
		}
	}

	//wg.Wait()

	cnt := CountSetBits(bitsArr)
	fmt.Println(cnt)
}

func processChunk(chunk []byte, bits []uint64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	var start = 0 // loop over every byte in chunk

	for i := 0; i < len(chunk); i++ {
		if chunk[i] == '\n' {
			// slice header only, no new backing array
			ipAddress := chunk[start:i:i]

			parseIPAndSet(ipAddress, bits, mutex)

			start = i + 1 // move past the '\n'
		}
	}

	// any final line without a trailing '\n'
	if start < len(chunk) {
		ipAddress := chunk[start:]

		parseIPAndSet(ipAddress, bits, mutex)
	}
}

func parseIPAndSet(ipAddress []byte, bits []uint64, mutex *sync.Mutex) {
	idx, err := parseIPv4(ipAddress)

	if err != nil {
		log.Println("Invalid IP address: ", string(ipAddress))

		return
	}

	SetBit(bits, idx, mutex)
}

// CountSetBits returns how many bits are 1 in arr.
func CountSetBits(arr []uint64) int {
	total := 0

	for _, w := range arr {
		total += bits.OnesCount64(w)
	}

	return total
}

var peakAlloc uint64

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// update peakAlloc if current Alloc is higher
	if m.Alloc > peakAlloc {
		peakAlloc = m.Alloc
	}

	fmt.Printf("Alloc = %v MiB\tPeakAlloc = %v MiB\n",
		m.Alloc/1024/1024,
		peakAlloc/1024/1024,
	)
}
func parseIPv4(line []byte) (uint32, error) {
	var ip, octet uint32
	dots := 0
	for _, c := range line {
		switch {
		case c >= '0' && c <= '9':
			octet = octet*10 + uint32(c-'0')
			if octet > 255 {
				return 0, fmt.Errorf("octet >255")
			}
		case c == '.':
			ip = (ip << 8) | octet
			octet = 0
			dots++
			if dots > 3 {
				return 0, fmt.Errorf("too many dots")
			}
		default:
			return 0, fmt.Errorf("invalid char %q", c)
		}
	}
	if dots != 3 {
		return 0, fmt.Errorf("wrong dot count")
	}
	ip = (ip << 8) | octet
	return ip, nil
}
