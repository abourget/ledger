package journal

import (
	"errors"
	"math/big"
	"strings"

	"github.com/abourget/ledger/parse"
)

var ErrInvalidAmount = errors.New("Unexpected type for amount given")

type Transaction struct {
	Node *parse.XactNode
}

func (tx *Transaction) Posting(account string) *Posting {
	for _, n := range tx.Node.Postings {
		if n.Account == account {
			return &Posting{n, tx}
		}
	}
	return nil
}

func (tx *Transaction) Postings() []*Posting {
	ps := make([]*Posting, len(tx.Node.Postings))
	for i, n := range tx.Node.Postings {
		ps[i] = &Posting{n, tx}
	}
	return ps
}

func (tx *Transaction) NewPosting(account string) *Posting {
	n := &parse.PostingNode{NodeType: parse.NodePosting}
	n.Account = account
	tx.Node.Postings = append(tx.Node.Postings, n)
	return &Posting{n, tx}
}

func (tx *Transaction) ImplicitAmount() *Amount {
	var amount *Amount
	for _, n := range tx.Node.Postings {
		if n.Amount != nil {
			a := nodeToAmount(n.Amount)
			if amount != nil {
				if amount.Commodity != a.Commodity {
					return nil
				}
				amount.Quantity.Add(amount.Quantity, a.Quantity)
			} else {
				amount = a
			}
		}
	}

	amount.Quantity.Neg(amount.Quantity)
	return amount
}

type Posting struct {
	Node        *parse.PostingNode
	Transaction *Transaction
}

func (p *Posting) Account() string {
	return strings.Trim(p.Node.Account, "()")
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
	if p.Node.Amount == nil {
		return p.Transaction.ImplicitAmount()
	}
	return nodeToAmount(p.Node.Amount)
}

func nodeToAmount(n *parse.AmountNode) *Amount {
	quant, ok := big.NewRat(0, 1).SetString(n.Quantity)
	if n.Negative {
		quant.Neg(quant)
	}
	if !ok {
		panic("cannot parse quantity: " + n.Quantity)
	}
	return &Amount{n.Commodity, quant}
}
