package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/bits"
	"net"
	"os"
	"strings"
	"sync"
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

	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	const chunkSize = 40 // e.g. 4 KB per chunk
	var leftover []byte  // carry any partial line
	buf := make([]byte, chunkSize)
	bitsArr := make([]uint64, wordsNeeded)

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
				wg.Add(1)
				go processChunk(full, bitsArr, &mutex, &wg)

				// leftover is any bytes after that newline
				leftover = data[cut+1:]
			}
		}

		if err != nil {
			if err == io.EOF {
				fmt.Println(len(leftover))
				// If the file does not end with a newline.
				if len(leftover) > 0 {
					wg.Add(1)
					processChunk(leftover, bitsArr, &mutex, &wg)
				}

				break
			}

			log.Fatal(err)
		}
	}

	//for i := 0; i < totalBits; i++ {
	//	if TestBit(bits, uint32(i)) {
	//		log.Println("Bit", i, "is ON", indexToIP(uint32(i)))
	//	}
	//}

	wg.Wait()

	cnt := CountSetBits(bitsArr)
	fmt.Println(cnt)
}

// processChunk handles a chunk of complete lines.
// Here we just print how many lines and bytes we got.
func processChunk(chunk []byte, bits []uint64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	lines := bytes.Count(chunk, []byte{'\n'})
	fmt.Printf("Got %d lines (%d bytes)\n", lines, len(chunk))
	//fmt.Println(string(chunk))

	ipAddresses := strings.Fields(string(chunk))

	for _, ipAddress := range ipAddresses {
		ip := net.ParseIP(ipAddress)

		if ip == nil {
			log.Println("Invalid IP address: ", ipAddress)

			continue
		}

		// same result as (1<<24)*firstNumber + (1<<16)*secondNumber + (1<<8)*thirdNumber + fourthNumber
		// 0 ... 2^32-1
		bit := binary.BigEndian.Uint32(ip.To4())

		fmt.Println(ip.To4(), bit)
		SetBit(bits, bit, mutex)
	}

	wg.Done()
}

func indexToIP(i uint32) net.IP {
	// make a 4-byte buffer
	buf := make([]byte, 4)
	// write i into buf in big-endian order
	binary.BigEndian.PutUint32(buf, i)
	// buf is now [b0 b1 b2 b3], so this is a valid IPv4
	return buf
}

func GetSetBits(arr []uint64) []uint32 {
	// 1) First count total bits so we can pre-alloc result slice
	var total int
	for _, w := range arr {
		total += bits.OnesCount64(w)
	}

	log.Println(total)
	res := make([]uint32, 0, total)

	// 2) Scan each word
	for wordIdx, w := range arr {
		base := uint32(wordIdx) * 64
		for w != 0 {
			// find lowest set bit
			tz := bits.TrailingZeros64(w)
			// record its absolute index
			res = append(res, base+uint32(tz))
			// clear that bit
			w &^= 1 << tz
		}
	}
	return res
}

// CountSetBits returns how many bits are 1 in arr.
func CountSetBits(arr []uint64) int {
	total := 0

	for _, w := range arr {
		total += bits.OnesCount64(w)
	}

	return total
}
