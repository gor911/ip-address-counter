package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

func Run2() {
	file, err := os.Open("resources/ip-addresses.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
		time.Sleep(1 * time.Second)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
