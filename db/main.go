package db

import (
	"database/sql"
	"log"

	"github.com/astroband/stellar-parallel-catchup/config"
)

const gapsQuery = `
	SELECT ledgerseq + 1 AS gap_start, next_nr - 1 AS gap_end
	FROM (
		SELECT ledgerseq, LEAD(ledgerseq) OVER (ORDER BY ledgerseq) AS next_nr
		FROM ledgerheaders
	) nr
	WHERE ledgerseq + 1 <> next_nr
`

const minQuery = `SELECT MIN(ledgerseq) FROM ledgerheaders`
const maxQuery = `SELECT MAX(ledgerseq) FROM ledgerheaders`

// Gap Represents gap in database
type Gap struct {
	Start int
	End   int
}

// GetGaps Returns gaps in current Stellar database
func GetGaps() (r []Gap) {
	// Empty head
	min := queryValue(minQuery)
	if min != *config.MinLedger {
		r = append(r, Gap{*config.MinLedger, min})
	}

	// Gaps
	rows, err := config.DB.Query(gapsQuery)
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

		r = append(r, Gap{start, finish})
	}

	// Tail
	max := queryValue(maxQuery)
	if max != *config.MaxLedger {
		r = append(r, Gap{max, *config.MaxLedger})
	}

	return r
}

func queryValue(query string) int {
	var value int

	err := config.DB.QueryRow(query).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1
		}

		log.Fatal(err)
	}

	return value
}
