package cmd

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

func Run2() {
	file, err := os.Open("resources/ip_addresses")

	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	bits := make([]uint64, wordsNeeded)
	mutex := sync.Mutex{}

	for scanner.Scan() {
		ip := net.ParseIP(scanner.Text())

		if ip == nil {
			log.Println("Invalid IP address: ", scanner.Text())

			continue
		}

		bit := binary.BigEndian.Uint32(ip.To4())

		SetBit(bits, bit, &mutex)

		PrintMemUsage()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	cnt := CountSetBits(bits)
	fmt.Println(cnt)
}
