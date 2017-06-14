package journal

import (
	"bytes"
	"os"
	"time"

	"github.com/abourget/ledger/parse"
	"github.com/abourget/ledger/print"
)

type Journal struct {
	tree *parse.Tree

	IncludedJournals map[string]*Journal
}

func Open(path string) (*Journal, error) {
	t, err := parse.Parse(path)
	if err != nil {
		return nil, err
	}
	return NewFromTree(t), nil
}

func NewFromTree(tree *parse.Tree) *Journal {
	j := &Journal{
		tree:             tree,
		IncludedJournals: make(map[string]*Journal),
	}
	return j
}

func (j *Journal) Transactions() []*Transaction {
	txs := make([]*Transaction, 0)
	for _, n := range j.tree.Root.Nodes {
		if x, ok := n.(*parse.XactNode); ok {
			txs = append(txs, &Transaction{x})
		}
	}
	return txs
}

func (j *Journal) AddTransaction(date time.Time, note string) *Transaction {
	sn := &parse.SpaceNode{NodeType: parse.NodeSpace}
	sn.Space = "\n"

	n := &parse.XactNode{NodeType: parse.NodeXact}
	n.Date = date
	n.Note = note

	j.tree.Root.Nodes = append(j.tree.Root.Nodes, sn, n)
	return &Transaction{n}
}

func (j *Journal) Marshal() ([]byte, error) {
	printer := print.New(j.tree)
	buf := &bytes.Buffer{}
	err := printer.Print(buf)
	return buf.Bytes(), err
}

func (j *Journal) SaveTo(path string) error {
	by, err := j.Marshal()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(by)
	return err
}
