package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xconstruct/ledger-utils/journal"
	"github.com/xconstruct/ledger-utils/reports"
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
		txs, err := j.Transactions()
		must(err)
		bal := reports.Balance(txs)
		for _, acc := range bal.Accounts {
			for _, am := range acc.Amounts {
				fmt.Println(acc.Name, " ", am)
			}
		}
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
