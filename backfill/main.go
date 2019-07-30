package backfill

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"text/template"

	"github.com/astroband/stellar-parallel-catchup/config"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)

const (
	dbName = "stellar.sqlite"
)

var tables = []string{"ledgerheaders", "txhistory", "txfeehistory", "upgradehistory", "scphistory"}

// Backfill Represents Backfill instance
type Backfill struct {
	Start  int
	Count  int
	Ledger int
	Dir    string
	DbFile string
}

// New Backfill Constructor
func New(start int, count int) *Backfill {
	ledger := start + count - 1
	dir := path.Join(*config.WorkDir, fmt.Sprintf("%s-%s", strconv.Itoa(ledger), strconv.Itoa(count)))
	dbFile := path.Join(dir, dbName)

	return &Backfill{start, count, ledger, dir, dbFile}
}

// Do Backfill payload
func (b *Backfill) Do() {
	log.Println("stellar-core catchup", b.catchupString())

	conf := b.prepare()

	b.run("stellar-core", "--conf", conf, "new-db")
	b.run("stellar-core", "--conf", conf, "catchup", b.catchupString())

	log.Println("sqlite export / psql -c", b.catchupString())
	b.truncDatabase()
	b.infill()

	b.cleanup()
}

func (b *Backfill) createConfig() string {
	path := path.Join(b.Dir, "stellar-core.cfg")

	t, err := template.ParseFiles(*config.StellarConfigTemplate)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(f, struct {
		Ledger int
		Count  int
		DB     string
	}{
		b.Ledger,
		b.Count,
		dbName,
	})

	return path
}

func (b *Backfill) catchupString() string {
	return fmt.Sprintf("%s/%s", strconv.Itoa(b.Ledger), strconv.Itoa(b.Count))
}

func (b *Backfill) prepare() string {
	os.MkdirAll(b.Dir, os.ModePerm)
	return b.createConfig()
}

func (b *Backfill) cleanup() {
	os.RemoveAll(b.Dir)
}

func (b *Backfill) truncDatabase() {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared", b.DbFile))
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)

	for _, table := range tables {
		db.Exec(fmt.Sprintf("DELETE FROM %s WHERE ledgerseq=1", table))
		db.Exec(fmt.Sprintf("DELETE FROM %s WHERE ledgerseq > ?", table), b.Ledger)
		db.Exec(fmt.Sprintf("DELETE FROM %s WHERE ledgerseq < ?", table), b.Start)
	}

	// db.Exec("DROP TABLE IF EXISTS accountdata")
	// db.Exec("DROP TABLE IF EXISTS accounts")
	// db.Exec("DROP TABLE IF EXISTS ban")
	// db.Exec("DROP TABLE IF EXISTS offers")
	// db.Exec("DROP TABLE IF EXISTS peers")
	// db.Exec("DROP TABLE IF EXISTS publishqueue")
	// db.Exec("DROP TABLE IF EXISTS pubsub")
	// db.Exec("DROP TABLE IF EXISTS quoruminfo")
	// db.Exec("DROP TABLE IF EXISTS scpquorums")
	// db.Exec("DROP TABLE IF EXISTS storestate")
	// db.Exec("DROP TABLE IF EXISTS trustlines")
}

// TODO: Replace with pg_loader || own bin util
func (b *Backfill) infill() {
	for _, table := range tables {
		exportCmd := exec.Command("sqlite3", "-header", "-csv", b.DbFile, fmt.Sprintf("select * from %s", table))
		importCmd := exec.Command("psql", "-c", fmt.Sprintf(`\copy %s from stdin csv header;`, table), (*config.DatabaseURL).String())
		importCmd.Stdout = os.Stdout
		importCmd.Stderr = os.Stdout

		//stdout, err := exportCmd.StdoutPipe()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		err := exportCmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		err = importCmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		err = importCmd.Wait()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (b *Backfill) run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = b.Dir
	// cmd.Stdout = os.Stdout
	// cmd.Start()
	// err := cmd.Wait()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
