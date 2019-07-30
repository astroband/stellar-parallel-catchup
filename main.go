package main

import (
	"log"
	"os"

	"github.com/astroband/stellar-parallel-catchup/backfill"
	"github.com/astroband/stellar-parallel-catchup/config"
	"github.com/astroband/stellar-parallel-catchup/db"
	"github.com/gammazero/workerpool"
)

var pool = workerpool.New(*config.Concurrency)

func main() {
	log.Println("stellar-parallel-catchup", config.Version)

	gaps := db.GetGaps()

	if len(gaps) == 0 {
		log.Println("Nothing to catch up!")
		os.Exit(0)
	}

	for _, gap := range gaps {
		log.Println("Gap:", gap.Start, "->", gap.End, "=", gap.Size, ":", gap.Chunks, "chunks")
	}

	for _, gap := range gaps {
		submitChunks(gap)
		submitTail(gap)
	}

	pool.StopWait()
}

func submitChunks(gap db.Gap) {
	for n := 0; n < gap.Chunks; n++ {
		start := gap.Start + n*(*config.ChunkSize)
		count := *config.ChunkSize

		backfill := backfill.New(start, count)
		pool.Submit(func() { backfill.Do() })
	}
}

func submitTail(gap db.Gap) {
	if gap.Tail > 0 {
		start := gap.Start + gap.Chunks*(*config.ChunkSize)
		count := gap.Tail

		backfill := backfill.New(start, count)
		pool.Submit(func() { backfill.Do() })
	}
}
