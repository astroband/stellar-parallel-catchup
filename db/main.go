package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/astroband/stellar-parallel-catchup/config"
	"github.com/lib/pq"
)

const gapsQuery = `
	SELECT ledgerseq + 1 AS gap_start, next_nr - 1 AS gap_end
	FROM (
		SELECT ledgerseq, LEAD(ledgerseq) OVER (ORDER BY ledgerseq) AS next_nr
		FROM (
			SELECT ledgerseq FROM ledgerheaders WHERE ledgerseq BETWEEN $1 AND $2
		) limits
	) nr
	WHERE ledgerseq + 1 <> next_nr
`

const minMaxQuery = `SELECT MIN(ledgerseq), MAX(ledgerseq) FROM ledgerheaders WHERE ledgerseq BETWEEN $1 AND $2`
const cleanupQuery = `DELETE FROM %s WHERE ledgerseq BETWEEN $1 AND $2`

// Gap Represents gap in database
type Gap struct {
	Start  int
	End    int
	Size   int
	Chunks int
	Tail   int
}

// NewGap Initializes new Gap
func NewGap(start int, finish int) Gap {
	size := finish + 1 - start
	chunks := size / *config.ChunkSize
	tail := size % *config.ChunkSize

	return Gap{
		start,
		finish,
		size,
		chunks,
		tail,
	}
}

// GetGaps Returns gaps in current Stellar database
func GetGaps() (r []Gap) {
	// Gaps
	rows, err := config.DB.Query(gapsQuery, *config.MinLedger, *config.MaxLedger)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var start, finish int

	for rows.Next() {
		err := rows.Scan(&start, &finish)
		if err != nil {
			log.Fatal(err)
		}

		r = append(r, NewGap(start, finish))
	}

	min, max := queryMinMax()

	if min == nil && max == nil {
		r = append(r, NewGap(*config.MinLedger, *config.MaxLedger))
	} else {
		if min != nil && *min != *config.MinLedger {
			r = append(r, NewGap(*config.MinLedger, *min-1))
		}

		if max != nil && *max != *config.MaxLedger {
			r = append(r, NewGap(*max+1, *config.MaxLedger))
		}
	}

	return r
}

// Cleanup Removes part of history from core database before import
func Cleanup(table string, min int, max int) {
	_, err := config.DB.Exec(fmt.Sprintf(cleanupQuery, pq.QuoteIdentifier(table)), min, max)
	if err != nil {
		log.Fatal(err)
	}
}

func queryMinMax() (*int, *int) {
	var min, max *int

	err := config.DB.QueryRow(minMaxQuery, *config.MinLedger, *config.MaxLedger).Scan(&min, &max)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Fatal(err)
	}

	return min, max
}
