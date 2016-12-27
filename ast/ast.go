// Package ast represents the different types used to represent a syntax tree of a Ledger file
package ast

import (
	"go/token"
	"time"
)

// Node is an element in the abstract syntax tree.
type Node interface {
	node()
	Pos() token.Pos
}

func (File) node()        {}
func (Transaction) node() {}
func (Posting) node()     {}
func (Comment) node()     {}

type File struct {
	Node Node

	Directives []Directive
	Comment    string
}

type Directive struct {
	Node Node
}

type Alias struct {
	Node               Node
	SourceAccount      string
	DestinationAccount string
}

type ApplyAccount struct {
	Node Node
}

type Transaction struct {
	Node Node

	Payee   string
	Pending bool
	Cleared bool
	Code    string // check number, or transaction number or whatnot..

	Postings []Posting
}

type Posting struct {
	Node Node

	Account     string
	Amount      string
	LineComment Comment
	TailComment Comment
}

type Comment struct {
	Node Node

	Text string
}

type Price struct {
	Node Node

	Date   time.Time
	Symbol string
	Price  string
}
