// ledgerfmt pretty-prints ledger files.
package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
)

var writeOutput = flag.Bool("w", false, "Write back to input file")
var sortXacts = flag.Bool("sort", false, "Sort transactions by date")

func main() {
	flag.Parse()

	var source io.Reader
	source = os.Stdin
	var filename = "stdin"

	inFile := flag.Arg(0)
	if inFile != "" {
		sourceFile, err := os.Open(inFile)
		if err != nil {
			log.Fatalln("Couldn't open source file:", err)
		}

		source = sourceFile
		filename = inFile
		defer sourceFile.Close()
	}

	cnt, err := ioutil.ReadAll(source)
	if err != nil {
		log.Fatalln("Error reading input:", err)
	}

	t := parse.New(filename, string(cnt))
	err = t.Parse()
	if err != nil {
		log.Fatalln("Parsing error:", err)
	}

	if *sortXacts {
		sortByDate(t)
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

// Represents a single xact plus whatever non-xact nodes come before it (eg,
// comments).
type xactWithPrefix struct {
	prefix []parse.Node
	xact   *parse.XactNode
}

// Sorts the nodes in t in-place by xact date, with non-xact notes staying in
// front of whatever xact they come before.
func sortByDate(t *parse.Tree) {
	// Lump in all non-xact nodes with the xact after them.
	var xacts []*xactWithPrefix
	var prefix []parse.Node
	for _, n := range t.Root.Nodes {
		if xact, ok := n.(*parse.XactNode); ok {
			xacts = append(xacts, &xactWithPrefix{prefix: prefix, xact: xact})
			prefix = nil
		} else {
			prefix = append(prefix, n)
		}
	}
	// Note that this leaves some trailing nodes in "prefix". We'll deal with them
	// later.

	sort.SliceStable(xacts, func(i, j int) bool {
		return xacts[i].xact.Date.Before(xacts[j].xact.Date)
	})

	t.Root.Nodes = nil
	for _, x := range xacts {
		t.Root.Nodes = append(t.Root.Nodes, x.prefix...)
		t.Root.Nodes = append(t.Root.Nodes, x.xact)
	}
	// Deal with any non-xacts at the end of the file.
	t.Root.Nodes = append(t.Root.Nodes, prefix...)
}
