package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/xconstruct/ledger/parse"
	"github.com/xconstruct/ledger/print"
)

var writeOutput = flag.Bool("w", false, "Write back to input file")

func main() {
	flag.Parse()

	var source io.Reader
	source = os.Stdin

	inFile := flag.Arg(0)
	if inFile != "" {
		sourceFile, err := os.Open(inFile)
		if err != nil {
			log.Fatalln("Couldn't open source file:", err)
		}

		source = sourceFile
		defer sourceFile.Close()
	}

	cnt, err := ioutil.ReadAll(source)
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

	var dest io.Writer
	dest = os.Stdout
	if inFile != "" && *writeOutput {
		destFile, err := os.Create(inFile)
		if err != nil {
			log.Fatalln("Couldn't write to file:", inFile)
		}
		dest = destFile
		defer destFile.Close()
	}

	_, err = dest.Write(buf.Bytes())
	if err != nil {
		log.Fatalln("Error writing to file:", err)

	}
}
