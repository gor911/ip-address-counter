package main

import (
	"fmt"
	"ip-address/cmd"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func main() {
	defer timer("main")()
	defer PrintMemUsage()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cmd.Run()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

// timer returns a function that prints the name argument and
// the elapsed time between the call to timer and the call to
// the returned function. The returned function is intended to
// be used in a defer statement:
//
//	defer timer("sum")()
func timer(name string) func() {
	start := time.Now()

	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
