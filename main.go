package main

import (
	"fmt"

	"github.com/astroband/stellar-parallel-catchup/db"
)

func main() {
	gaps := db.GetGaps()

	for _, gap := range gaps {
		fmt.Println(gap)
	}
}
