package main

import (
	"fmt"
	"log"

	"github.com/astroband/stellar-parallel-catchup/config"
	"github.com/astroband/stellar-parallel-catchup/db"
	"github.com/gammazero/workerpool"
)

var pool = workerpool.New(*config.Concurrency)

func main() {
	log.Println("stellar-parallel-catchup ", config.Version)

	gaps := db.GetGaps()

	for index, gap := range gaps {
		fmt.Println(index, gap.Chunks, gap.Tail)
	}

	pool.StopWait()
}
