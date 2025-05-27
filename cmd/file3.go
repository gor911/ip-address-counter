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
	"runtime"
	"strings"
	"sync"
)

func Run3() {
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

	var leftover []byte // carry any partial line
	buf := make([]byte, chunkSize)
	bitsArr := make([]uint64, wordsNeeded)

	for {
		n, err := f.Read(buf)

		if n > 0 {
			// 1) combine leftover from last read with new bytes
			data := append(leftover, buf[:n]...)

			PrintMemUsage()

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
				processChunk(full, bitsArr, &mutex, &wg)

				// leftover is any bytes after that newline
				leftover = data[cut+1:]
			}
		}

		if err != nil {
			if err == io.EOF {
				fmt.Println("leftover", len(leftover))
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

	wg.Wait()

	cnt := CountSetBits(bitsArr)
	fmt.Println(cnt)
}

// processChunk handles a chunk of complete lines.
// Here we just print how many lines and bytes we got.
func processChunk(chunk []byte, bits []uint64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

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

		SetBit(bits, bit, mutex)
	}
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
	//fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	//fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	//fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	//fmt.Printf("\tNumGC = %v\n", m.NumGC)

	// update peakAlloc if current Alloc is higher
	if m.Alloc > peakAlloc {
		peakAlloc = m.Alloc
	}

	fmt.Printf("Alloc = %v MiB\tPeakAlloc = %v MiB\n",
		m.Alloc/1024/1024,
		peakAlloc/1024/1024,
	)
}
