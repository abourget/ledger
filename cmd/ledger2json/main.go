package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/xconstruct/ledger/parse"
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

	out, err := json.MarshalIndent(t.Root, "", "  ")
	if err != nil {
		log.Fatalln("json encoding:", err)
	}

	os.Stdout.Write(out)
}
