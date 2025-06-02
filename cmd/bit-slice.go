package cmd

import (
	"sync"
)

// totalBits is 2^32 bits
const totalBits = 1 << 32

// we pack 64 bits per uint64 word
const wordBits = 64

// wordsNeeded = totalBits/wordBits = 1<<(32-6) = 1<<26
const wordsNeeded = totalBits / wordBits

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
