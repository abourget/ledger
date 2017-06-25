// Package journal provides a simple interface for accessing ledger files.
package journal

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/xconstruct/ledger/parse"
	"github.com/xconstruct/ledger/print"
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

func (j *Journal) Transactions() ([]*Transaction, error) {
	txs := make([]*Transaction, 0)
	for _, n := range j.tree.Root.Nodes {
		if x, ok := n.(*parse.XactNode); ok {
			txs = append(txs, &Transaction{x})
		}

		if d, ok := n.(*parse.DirectiveNode); ok && d.Directive == "include" {
			inc, err := j.IncludeJournal(d.Args)
			if err != nil {
				return txs, err
			}
			incTxs, err := inc.Transactions()
			if err != nil {
				return txs, err
			}
			txs = append(txs, incTxs...)
		}
	}
	return txs, nil
}

func (j *Journal) IncludeJournal(path string) (*Journal, error) {
	path = filepath.Join(j.tree.FileName, "..", path)
	if inc, ok := j.IncludedJournals[path]; ok {
		return inc, nil
	}

	inc, err := Open(path)
	if err != nil {
		return nil, err
	}
	j.IncludedJournals[path] = inc
	return inc, nil
}

func (j *Journal) AddTransaction(date time.Time, desc string) *Transaction {
	sn := &parse.SpaceNode{NodeType: parse.NodeSpace}
	sn.Space = "\n"

	n := &parse.XactNode{NodeType: parse.NodeXact}
	n.Date = date
	n.Description = desc

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
