// Ledger-Go is a Go implementation of the Ledger cli tool.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/xconstruct/ledger/journal"
	"github.com/xconstruct/ledger/tools/reports"
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

	if *fname == "" {
		*fname = os.Getenv("LEDGER_FILE")
		if *fname == "" {
			fmt.Println("Please specify an existing journal file with -f or LEDGER_FILE")
			os.Exit(1)
		}
	}

	j, err := journal.Open(*fname)
	must(err)

	switch {
	case cmd == "balance" || cmd == "bal":
		filter := regexp.MustCompile("(?i)" + flag.Arg(1))

		txs, err := j.Transactions()
		must(err)
		bal := reports.BalanceFiltered(txs, func(acc string) bool {
			return filter.MatchString(acc)
		})
		must(bal.Print(os.Stdout))
	case cmd == "testadd":
		tx := j.AddTransaction(time.Now(), "This is a test transaction")
		tx.NewPosting("Expenses:Testing").SetAmount("EUR", 120)
		tx.NewPosting("Expenses:OtherTesting").SetAmount("EUR", -120)
		must(j.SaveTo(*fname))
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
