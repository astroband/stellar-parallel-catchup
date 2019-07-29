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
		FROM (
			SELECT ledgerseq FROM ledgerheaders WHERE ledgerseq BETWEEN $1 AND $2
		) limits
	) nr
	WHERE ledgerseq + 1 <> next_nr
`

const minQuery = `SELECT MIN(ledgerseq) FROM ledgerheaders WHERE ledgerseq >= $1`
const maxQuery = `SELECT MAX(ledgerseq) FROM ledgerheaders WHERE ledgerseq <= $1`

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
	size := finish - start
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

	min := queryValue(minQuery, *config.MinLedger)
	max := queryValue(maxQuery, *config.MaxLedger)

	if min == -1 && max == -1 {
		r = append(r, NewGap(*config.MinLedger, *config.MaxLedger))
	} else {
		if min != -1 {
			r = append(r, NewGap(*config.MinLedger, min))
		}

		if max != -1 {
			r = append(r, NewGap(max, *config.MinLedger))
		}
	}

	//return []Gap{NewGap(25069000, 25069442)}

	//return []Gap{NewGap(2, 1000), NewGap(5000, 6000)}

	return r
}

func queryValue(query string, param int) int {
	var value *int

	err := config.DB.QueryRow(query, param).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1
		}

		log.Fatal(err)
	}

	if value == nil {
		return -1
	}

	return *value
}
