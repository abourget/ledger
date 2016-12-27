package ledger

import (
	"github.com/abourget/ledger/ast"
	"github.com/abourget/ledger/parser"
)

type Ledger struct {
	*ast.Node
	Commodities  []Commodity
	Transactions []Transaction
	Prices       []Price
	Nodes        []ast.Node
}

type Commodity struct {
	*ast.Node
}

type Transaction struct {
	*ast.Node
	Comment  string
	Postings []Posting
}

type Posting struct {
	*ast.Node
	Comment string
	Account string
	Amount  Money
}

type Money struct {
	*ast.Node
	FloatAmount float64
	Amount      int
	Decimal     int
	Currency    string
}

func parse(in []byte) (*ast.File, error) {
	return parser.Parse(in)
}
