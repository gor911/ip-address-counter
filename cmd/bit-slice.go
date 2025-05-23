package cmd

import (
	"fmt"
	"sync"
	"time"
)

// totalBits is 2^32 bits
const totalBits = 1 << 32

// we pack 64 bits per uint64 word
const wordBits = 64

// wordsNeeded = totalBits/wordBits = 1<<(32-6) = 1<<26
const wordsNeeded = totalBits / wordBits

func RunBitSlice() {
	// Allocate the slice. Under the hood this uses exactly 512 MiB.
	bits := make([]uint64, wordsNeeded)
	mutex := sync.Mutex{}
	//bits := new([wordsNeeded]uint64)

	// Example index (must be < 2^32)

	for i := 0; i < 10000; i++ {
		go func(mutex *sync.Mutex) {
			//mutex.Lock()
			SetBit(bits, uint32(i), mutex)
			//mutex.Unlock()
		}(&mutex)
	}

	var ii uint32 = 123456789

	// Test bit i
	if TestBit(bits, ii) {
		fmt.Println("Bit", ii, "is ON")
	} else {
		fmt.Println("Bit", ii, "is OFF")
	}

	// Clear bit i
	ClearBit(bits, ii)
	fmt.Println("After clear:", TestBit(bits, ii)) // should be false

	time.Sleep(5 * time.Second)
}

// SetBit turns on bit i (0 ≤ i < 2^32)
func SetBit(arr []uint64, i uint32, mutex *sync.Mutex) {
	idx := i >> 6             // which uint64 word holds bit i, same as i/64
	pos := i & (wordBits - 1) // which bit inside that word, same as i%64
	//mutex.Lock()
	arr[idx] |= 1 << pos // use OR to turn on exactly that bit
	//mutex.Unlock()
}

// ClearBit turns off bit i
func ClearBit(arr []uint64, i uint32) {
	idx := i >> 6
	pos := i & (wordBits - 1)
	arr[idx] &^= 1 << pos
}

// TestBit returns true if bit i is 1
func TestBit(arr []uint64, i uint32) bool {
	idx := i >> 6
	pos := i & (wordBits - 1)
	return arr[idx]&(1<<pos) != 0
}
