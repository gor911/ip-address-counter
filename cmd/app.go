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

var peakAlloc uint64
var memStats runtime.MemStats

func calcPeakAlloc(m *runtime.MemStats) {
	runtime.ReadMemStats(m)

	// update peakAlloc if current Alloc is higher
	if m.Alloc > peakAlloc {
		peakAlloc = m.Alloc
	}
}

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

	defer f.Close()

	// Allocate a single read‐buffer of capacity chunkSize+16.
	// We keep len(fullData) = how many bytes are “leftover” from previous read.
	// At most we’ll ever have chunkSize + 15 bytes in it (since no IP line exceeds 16 bytes).
	fullData := make([]byte, 0, chunkSize+16)

	// Allocate the 2^32‐bit bitset = (2^32 / 64) = 1<<26 uint64 words = 512 MiB.
	bitsArr := make([]uint64, wordsNeeded)

	for {
		// Compute how many bytes are currently in fullData
		leftoverLen := len(fullData)

		// Read exactly chunkSize bytes into fullData[leftoverLen : leftoverLen+chunkSize]
		// (guaranteed not to exceed cap because leftoverLen ≤ 15)
		readBuf := fullData[leftoverLen : leftoverLen+chunkSize]
		n, err := f.Read(readBuf)

		if n > 0 {
			// Extend fullData’s length to include the n bytes just read
			fullData = fullData[:leftoverLen+n]

			// Find the last '\n' in fullData. Everything up to that newline is (complete) lines.
			if cut := bytes.LastIndexByte(fullData, '\n'); cut >= 0 {
				// Process the chunk of whole lines (no trailing newline, so lines aren’t split).
				processChunk(fullData[:cut], bitsArr)
				calcPeakAlloc(&memStats)

				// Move any partial line (≤ 15 bytes) back to front
				rem := len(fullData) - (cut + 1)        // ≤ 15
				copy(fullData[0:rem], fullData[cut+1:]) // move leftover after the '\n' to start
				fullData = fullData[:rem]               // shrink length to just the leftover
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

	calcPeakAlloc(&memStats)
	fmt.Printf("Peak memory = %d MiB\n", peakAlloc/1024/1024)

	//Count all set bits in bitsArr:
	count := CountSetBits(bitsArr)
	fmt.Println("Total unique IPs:", count)
}

// processChunk parses every line in chunk (each line is "a.b.c.d") and sets its bit.
func processChunk(chunk []byte, bitsArr []uint64) {
	start := 0
	for i := 0; i < len(chunk); i++ {
		if chunk[i] == '\n' {
			// slice header only, no new backing array:
			line := chunk[start:i:i] // len = i-start, cap = i-start

			parseIPAndSet(line, bitsArr)
			start = i + 1
		}
	}
	// If the final line has no trailing '\n', handle it:
	if start < len(chunk) {
		line := chunk[start:]
		parseIPAndSet(line, bitsArr)
	}
}

// parseIPAndSet turns "a.b.c.d" (as bytes) into a uint32 and sets that bit.
func parseIPAndSet(ipBytes []byte, bitsArr []uint64) {
	idx, err := parseIPv4(ipBytes)
	if err != nil {
		// If invalid, you can log or ignore. Converting to string(ipBytes) would allocate,
		// so we just skip printing the offending line.
		return
	}
	// Set bit idx in the 2^32-bit array:
	wordIdx := idx >> 6            // idx / 64
	bitPos := idx & (wordBits - 1) // idx % 64
	bitsArr[wordIdx] |= 1 << bitPos
}

// CountSetBits returns how many bits are 1 in arr.
func CountSetBits(arr []uint64) int {
	total := 0

	for _, w := range arr {
		total += bits.OnesCount64(w)
	}

	return total
}

// parseIPv4 parses "a.b.c.d" in-place (no allocations) and returns a uint32 index.
func parseIPv4(line []byte) (uint32, error) {
	var ip, octet uint32
	dots := 0

	for _, c := range line {
		switch {
		case c >= '0' && c <= '9':
			octet = octet*10 + uint32(c-'0')
			if octet > 255 {
				return 0, fmt.Errorf("octet > 255")
			}
		case c == '.':
			ip = (ip << 8) | octet
			octet = 0
			dots++
			if dots > 3 {
				return 0, fmt.Errorf("too many dots in %q", line)
			}
		default:
			return 0, fmt.Errorf("invalid char %q in %q", c, line)
		}
	}
	if dots != 3 {
		return 0, fmt.Errorf("wrong dot count in %q", line)
	}
	ip = (ip << 8) | octet
	return ip, nil
}
