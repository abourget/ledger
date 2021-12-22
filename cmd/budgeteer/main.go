// Budgeteer is an experimental monthly budget calculator.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abourget/ledger/journal"
	"github.com/abourget/ledger/tools/budget"
)

var fname = flag.String("f", "", "ledger file")

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	flag.Parse()
	cmd := flag.Arg(0)

	j, err := journal.Open(*fname)
	must(err)

	switch {
	case cmd == "balance" || cmd == "bal":
		bal, err := budget.Balance(j, time.Time{})
		must(err)
		must(bal.Print(os.Stdout))
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
