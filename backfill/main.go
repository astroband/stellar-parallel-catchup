package backfill

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"text/template"

	"github.com/astroband/stellar-parallel-catchup/config"
)

// Backfill Represents Backfill instance
type Backfill struct {
	Start  int
	Count  int
	Ledger int
}

// New Backfill Constructor
func New(start int, count int) *Backfill {
	return &Backfill{start, count, start + count}
}

// Do Backfill payload
func (b *Backfill) Do() {
	path := b.createConfig()
	catchup := fmt.Sprintf("%s/%s", strconv.Itoa(b.Ledger), strconv.Itoa(b.Count))

	log.Println("Catching up", catchup)

	run("stellar-core", "--conf", path, "new-db")
	run("stellar-core", "--conf", path, "catchup", catchup)
}

func (b *Backfill) createConfig() string {
	path := path.Join(
		*config.WorkDir,
		fmt.Sprintf("stellar-core-%s-%s.cfg", strconv.Itoa(b.Ledger), strconv.Itoa(b.Count)),
	)

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
	}{
		b.Ledger,
		b.Count,
	})

	return path
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
