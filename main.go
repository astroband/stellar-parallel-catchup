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

	setMaxLedger()
	gaps := getGaps()

	for _, gap := range gaps {
		log.Println("Gap:", gap.Start, "->", gap.End, "=", gap.Size, ":", gap.Chunks+1, "chunks")
	}

	for _, gap := range gaps {
		submitChunks(gap)
		submitTail(gap)
	}

	pool.StopWait()
}

func setMaxLedger() {
	max := db.GetMaxLedger()
	if max == nil || *max == 1 {
		log.Fatal("Can not catch up fresh database. Run stellar-core node and wait for initial catchup to finish!")
	}

	if *max < *config.MaxLedger {
		log.Fatal("Can not catchup segment after the last ledger database has, set --maxLedger to any ledger less than the last")
	}

	if *config.MaxLedger == 0 {
		*config.MaxLedger = *max
	}

	if *config.MinLedger > *config.MaxLedger {
		log.Fatal("Can not catchup segment after the last ledger database has.")
	}

	log.Printf("Analysing database from ledger %v to %v", *config.MinLedger, *config.MaxLedger)
}

func getGaps() []db.Gap {
	gaps := db.GetGaps()
	if len(gaps) == 0 {
		log.Println("Nothing to catch up, database consistent!")
		os.Exit(0)
	}
	return gaps
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
