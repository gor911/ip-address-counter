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
	"unsafe"
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

	//var leftover []byte // carry any partial line
	leftover := make([]byte, 0, chunkSize+32)

	buf := make([]byte, chunkSize)
	//data := make([]byte, chunkSize*2)
	//fmt.Println(cap(data))
	bitsArr := make([]uint64, wordsNeeded)
	//fmt.Println(cap(data))
	//fmt.Println(len(data))

	for {
		n, err := f.Read(buf)

		if n > 0 {
			//PrintMemUsage()
			//data = append(data, buf[:n]...)
			//data = buf[:n]
			//fmt.Println(cap(data))

			//data = data[:0]
			//fmt.Println(cap(data))

			//fmt.Println(len(data))

			// 2) Append into the same slice, reusing capacity
			leftover = append(leftover, buf[:n]...)
			//fmt.Println(uintptr(unsafe.Pointer(&leftover[0])))
			//
			//fmt.Println(&leftover[0])
			//fmt.Println(cap(leftover))
			//PrintMemUsage()

			if cut := bytes.LastIndexByte(leftover, '\n'); cut >= 0 {
				//wg.Add(1)
				//processChunk(leftover[:cut], bitsArr, &mutex, &wg)

				//processLine(leftover[:cut], bitsArr, &mutex, &wg)
				processChunk2(leftover[:cut], bitsArr, &mutex, &wg)

				copy(leftover, leftover[cut+1:])
				leftover = leftover[:len(leftover)-cut-1]
				//fmt.Println(cap(leftover))

				//fmt.Println(len(leftover))
			}

			PrintMemUsage()

			//lines := bytes.Split(buf[:n], []byte("\n"))
			//
			//fmt.Println(len(lines))

			//wg.Add(1)
			//processLine(lines, bitsArr, &mutex, &wg)

			// Save last part as leftover
			//leftover = lines[len(lines)-1]
		}

		if err != nil {
			if err == io.EOF {
				fmt.Println("leftover")
				fmt.Println("leftover", len(leftover))
				//// If the file does not end with a newline.
				//if len(leftover) > 0 {
				//	wg.Add(1)
				//	processChunk(leftover, bitsArr, &mutex, &wg)
				//}

				break
			}

			log.Fatal(err)
		}
	}

	//wg.Wait()

	cnt := CountSetBits(bitsArr)
	fmt.Println(cnt)
}

// processChunk handles a chunk of complete lines.
// Here we just print how many lines and bytes we got.
func processChunk(chunk []byte, bits []uint64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	//defer wg.Done()

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

func processLine(chunk []byte, bits []uint64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	//defer wg.Done()

	ipAddresses := bytes.Split(chunk, []byte("\n"))

	for _, ipAddress := range ipAddresses {
		ip := net.ParseIP(string(ipAddress))

		if ip == nil {
			log.Println("Invalid IP address: ", string(ipAddress))

			continue
		}

		// same result as (1<<24)*firstNumber + (1<<16)*secondNumber + (1<<8)*thirdNumber + fourthNumber
		// 0 ... 2^32-1
		processIP(ip.To4(), bits, mutex)
	}
}

func processIP(ip []byte, bits []uint64, mutex *sync.Mutex) {
	bit := binary.BigEndian.Uint32(ip)

	SetBit(bits, bit, mutex)
}

func processChunk2(chunk []byte, bits []uint64, mutex *sync.Mutex, wg *sync.WaitGroup) {
	var start = 0 // loop over every byte in chunk

	for i := 0; i < len(chunk); i++ {
		if chunk[i] == '\n' {
			// slice header only, no new backing array
			ipAddress := chunk[start:i:i]

			//idx, err := parseIPv4(ipAddress)
			//if err == nil {
			//	SetBit(bits, idx, mutex)
			//}

			ipAddressStr := *(*string)(unsafe.Pointer(&ipAddress))
			ip := net.ParseIP(ipAddressStr) // still allocates the IP slice
			//PrintMemUsage()

			//ip := net.ParseIP(string(ipAddress))
			//
			if ip == nil {
				log.Println("Invalid IP address: ", string(ipAddress))

				continue
			}

			processIP(ip.To4(), bits, mutex)

			start = i + 1 // move past the '\n'
		}
	}

	// any final line without a trailing '\n'
	if start < len(chunk) {
		//ipAddress := chunk[start:]
		//
		//ip := net.ParseIP(string(ipAddress))
		//
		//if ip == nil {
		//	log.Println("Invalid IP address: ", string(ipAddress))
		//
		//	return
		//}
		//
		//processIP(ip.To4(), bits, mutex) // your IP‐handling function
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
