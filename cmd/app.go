package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/bits"
	"os"
	"runtime"
)

const (
	totalBits   = 1 << 32 // 2^32 bits
	wordBits    = 64      // pack 64 bits per uint64 word
	wordsNeeded = totalBits / wordBits
	chunkSize   = 10 * 1024 * 1024 // 10 MiB
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

	var memStats runtime.MemStats
	fullData := make([]byte, 0, chunkSize+16)
	buf := make([]byte, chunkSize)
	bitsArr := make([]uint64, wordsNeeded)

	for {
		n, err := f.Read(buf)

		if n > 0 {
			fullData = append(fullData, buf[:n]...)

			if cut := bytes.LastIndexByte(fullData, '\n'); cut >= 0 {
				processChunk(fullData[:cut], bitsArr)
				calcPeakAlloc(&memStats)

				copy(fullData, fullData[cut+1:])
				fullData = fullData[:len(fullData)-cut-1]
			}
		}

		if err != nil {
			if err == io.EOF {
				// If the file does not end with a newline.
				if len(fullData) > 0 {
					processChunk(fullData, bitsArr)
				}

				break
			}

			log.Fatal(err)
		}
	}

	count := CountSetBits(bitsArr)
	fmt.Println("Total unique IPs:", count)
}

func processChunk(chunk []byte, bits []uint64) {
	var start = 0 // loop over every byte in chunk

	for i := 0; i < len(chunk); i++ {
		if chunk[i] == '\n' {
			// slice header only, no new backing array
			ipAddress := chunk[start:i:i]

			parseIPAndSet(ipAddress, bits)

			start = i + 1 // move past the '\n'
		}
	}

	// any final line without a trailing '\n'
	if start < len(chunk) {
		ipAddress := chunk[start:]

		parseIPAndSet(ipAddress, bits)
	}
}

func parseIPAndSet(ipAddress []byte, bits []uint64) {
	idx, err := parseIPv4(ipAddress)

	if err != nil {
		log.Println("Invalid IP address: ", string(ipAddress))

		return
	}

	SetBit(bits, idx)
}

// CountSetBits returns how many bits are 1 in arr.
func CountSetBits(arr []uint64) int {
	total := 0

	for _, w := range arr {
		total += bits.OnesCount64(w)
	}

	return total
}

// SetBit turns on bit i (0 ≤ i < 2^32)
func SetBit(arr []uint64, i uint32) {
	idx := i >> 6             // which uint64 word holds bit i, same as i/64
	pos := i & (wordBits - 1) // which bit inside that word, same as i%64
	arr[idx] |= 1 << pos      // use OR to turn on exactly that bit
}

var peakAlloc uint64

func calcPeakAlloc(m *runtime.MemStats) {
	runtime.ReadMemStats(m)

	// update peakAlloc if current Alloc is higher
	if m.Alloc > peakAlloc {
		peakAlloc = m.Alloc
	}
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
