package journal

import (
	"errors"
	"math/big"

	"github.com/abourget/ledger/parse"
)

var ErrInvalidAmount = errors.New("Unexpected type for amount given")

type Transaction struct {
	Node *parse.XactNode
}

func (tx *Transaction) Posting(account string) *Posting {
	for _, n := range tx.Node.Postings {
		if n.Account == account {
			return &Posting{Node: n}
		}
	}
	return nil
}

func (tx *Transaction) Postings() []*Posting {
	ps := make([]*Posting, len(tx.Node.Postings))
	for i, n := range tx.Node.Postings {
		ps[i] = &Posting{Node: n}
	}
	return ps
}

func (tx *Transaction) NewPosting(account string) *Posting {
	n := &parse.PostingNode{NodeType: parse.NodePosting}
	n.Account = account
	tx.Node.Postings = append(tx.Node.Postings, n)
	return &Posting{n}
}

type Posting struct {
	Node *parse.PostingNode
}

func (p *Posting) SetAmount(commodity string, amount interface{}) error {
	v := amountToString(amount)
	if v == "" {
		return ErrInvalidAmount
	}

	if p.Node.Amount == nil {
		p.Node.Amount = &parse.AmountNode{NodeType: parse.NodeAmount}
	}
	p.Node.Amount.Raw = ""
	p.Node.Amount.Commodity = commodity
	p.Node.Amount.Quantity = v
	return nil
}

func (p *Posting) Amount() *Amount {
	quant, ok := big.NewRat(0, 1).SetString(p.Node.Amount.Quantity)
	if p.Node.Amount.Negative {
		quant.Neg(quant)
	}
	if !ok {
		panic("cannot parse quantity: " + p.Node.Amount.Quantity)
	}
	return &Amount{p.Node.Amount.Commodity, quant}
}
