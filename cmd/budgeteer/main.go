package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xconstruct/ledger/journal"
	"github.com/xconstruct/ledger/tools/budget"
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
		since := time.Date(2017, 6, 1, 0, 0, 0, 0, time.Local)
		bal, err := budget.Balance(j, since)
		must(err)
		for _, acc := range bal.Accounts {
			for _, am := range acc.Amounts {
				fmt.Println(acc.Name, " ", am)
			}
		}
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
