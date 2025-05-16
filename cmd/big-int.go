package cmd

import (
	"fmt"
	"math/big"
)

func RunBigInt() {
	// Create an empty big.Int (all bits = 0)
	bits := big.NewInt(0)

	for i := 0; i < wordsNeeded; i++ {
		bits.SetBit(bits, i, 1)
	}
	// Set bit at index 3 to 1

	// Check bit at index 3
	if bits.Bit(3) == 1 {
		fmt.Println("Bit 3 is ON")
	}

	// Clear bit 3 (set it back to 0)
	bits.SetBit(bits, 3, 0)
	fmt.Println("Bit 3 now:", bits.Bit(3)) // prints 0
}
