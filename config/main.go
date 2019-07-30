package config

import (
	"database/sql"
	"log"

	"gopkg.in/alecthomas/kingpin.v2"

	_ "github.com/lib/pq" // Database driver
)

var (
	// Version Application version
	Version = "0.1"

	// DB Database connection
	DB *sql.DB

	// DatabaseURL Stellar Core database URL
	DatabaseURL = kingpin.
			Flag("database-url", "Stellar Core database URL").
			Default("postgres://localhost/core?sslmode=disable").
			OverrideDefaultFromEnvar("DATABASE_URL").
			URL()

	// MinLedger Starting ledger to catch up from
	MinLedger = kingpin.
			Flag("min-ledger", "Minimal ledger to start from").
			Default("250690000").
			OverrideDefaultFromEnvar("MIN_LEDGER").
			Int()

	// MaxLedger Ledger to checkup to
	MaxLedger = kingpin.
			Flag("max-ledger", "Maximum ledger to finish on (is loaded from public Horizon by default)").
			Default("25069442").
			OverrideDefaultFromEnvar("MAX_LEDGER").
			Int()

	// ChunkSize Chunk size to download ledgers
	ChunkSize = kingpin.
			Flag("chunk-size", "Chunk size").
			Default("1000").
			OverrideDefaultFromEnvar("CHUNK_SIZE").
			Short('s').
			Int()

	// Concurrency Parallel streams
	Concurrency = kingpin.
			Flag("concurrency", "Concurrency").
			Default("5").
			OverrideDefaultFromEnvar("CONCURRENCY").
			Short('c').
			Int()

	// WorkDir Where to create temporary stellar files
	WorkDir = kingpin.
		Flag("work-dir", "Working directory").
		Default("/tmp").
		OverrideDefaultFromEnvar("WORK_DIR").
		String()

	// StellarConfigTemplate Path to stellar-core.cfg template file
	StellarConfigTemplate = kingpin.
				Flag("stellar-config-template", "stellar-core.cfg template file location").
				Default("./templates/stellar-core.cfg").
				OverrideDefaultFromEnvar("STELLAR_CONFIG_TEMPLATE").
				String()
)

func init() {
	kingpin.Version(Version)
	kingpin.Parse()

	initDB()
}

func initDB() {
	databaseDriver := (*DatabaseURL).Scheme

	db, err := sql.Open(databaseDriver, (*DatabaseURL).String())
	if err != nil {
		log.Fatal(err)
	}

	DB = db
}
