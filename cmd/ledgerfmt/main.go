package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
)

func main() {
	cnt, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}

	t := parse.New("stdin", string(cnt))
	err = t.Parse()
	if err != nil {
		log.Fatalln("Parsing error:", err)
	}

	printer := print.New(t)

	buf := &bytes.Buffer{}
	err = printer.Print(buf)
	if err != nil {
		log.Fatalln("rendering ledger file:", err)
	}

	os.Stdout.Write(buf.Bytes())
}
