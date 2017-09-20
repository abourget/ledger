package print

import (
	"bytes"
	"strings"
	"time"

	"github.com/xconstruct/ledger/parse"
)

func accountLength(post *parse.PostingNode) int {
	accountLen := len([]rune(post.Account))
	if post.IsCleared {
		accountLen += 2
	} else if post.IsPending {
		accountLen += 2
	}
	return accountLen
}

func quantityLength(amount *parse.AmountNode) int {
	quantityLen := len([]rune(amount.Quantity))
	if amount.Negative {
		quantityLen++
	}
	return quantityLen
}

func toDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func (p *Printer) commentReturns(node *parse.XactNode, input string) string {
	width := p.PostingsIndent
	if width == 0 {
		width = len(node.Postings[0].AccountPreSpace)
	}
	return strings.Replace(input, "\n", "\n"+strings.Repeat(" ", width), -1)
}

func (p *Printer) postingAccountPreSpace(node *parse.XactNode, post *parse.PostingNode) string {
	if p.PostingsIndent == 0 {
		return node.Postings[0].AccountPreSpace
	}
	return strings.Repeat(" ", p.PostingsIndent)
}

func (p *Printer) postingAccountPostSpace(node *parse.XactNode, post *parse.PostingNode) string {
	var longestAccountName int
	var longestQuantity int
	for _, post := range node.Postings {
		accountLen := accountLength(post)
		if accountLen > longestAccountName {
			longestAccountName = accountLen
		}
		if post.Amount != nil {
			quantityLen := quantityLength(post.Amount)
			if quantityLen > longestQuantity {
				longestQuantity = quantityLen
			}
		}
	}

	if longestAccountName < p.MinimumAccountWidth {
		longestAccountName = p.MinimumAccountWidth
	}

	accountLen := accountLength(post)
	baseSpacing := longestAccountName - accountLen + 4
	spaceFunc := func(spaces int) string {
		return strings.Repeat(" ", spaces)
	}

	// take longest width, add 2 spaces, if value is negative, add 2 more spaces..
	// right align on the AMOUNT, put the negative sign in front, and the commodity
	// AFTER.
	// if one is a ValueExpr, align with the left-most character.
	// if there's a BalanceAssignment, there align that left-most..
	// if there's no other amount (not price, not balanceassignment), then no space at all.
	if post.Amount != nil && post.Amount.ValueExpr != "" {
		return spaceFunc(baseSpacing)
	}
	if post.Amount == nil && post.BalanceAssignment == nil && post.LotPrice == nil && post.LotDate.IsZero() && post.Price == nil && post.Note == "" {
		return ""
	}

	if post.Amount != nil && post.Amount.Quantity != "" {
		baseSpacing += (longestQuantity - quantityLength(post.Amount))
	}

	return spaceFunc(baseSpacing)
}

func amount(amount *parse.AmountNode) (out string) {
	if amount.ValueExpr != "" {
		return amount.ValueExpr
	}
	if amount.Negative {
		out += "-"
	}
	if amount.Commodity == "$" {
		out += amount.Commodity + amount.Quantity
	} else {
		out += amount.Quantity
		if amount.Commodity != "" {
			out += " " + amount.Commodity
		}
	}
	return out
}

func (p *Printer) writePlainXact(b *bytes.Buffer, x *parse.XactNode) {
	b.WriteString(toDate(x.Date))
	if !x.EffectiveDate.IsZero() {
		b.WriteString(" = ")
		b.WriteString(toDate(x.EffectiveDate))
	}
	if x.IsPending {
		b.WriteString(" !")
	}
	if x.IsCleared {
		b.WriteString(" *")
	}
	if x.Code != "" {
		b.WriteString(" (")
		b.WriteString(x.Code)
		b.WriteString(")")
	}
	b.WriteByte(' ')
	b.WriteString(x.Description)
	if x.Note != "" {
		b.WriteString(x.NotePreSpace)
		b.WriteString(p.commentReturns(x, x.Note))
	}

	for _, posting := range x.Postings {
		b.WriteByte('\n')
		b.WriteString(p.postingAccountPreSpace(x, posting))
		if posting.IsPending {
			b.WriteString("! ")
		}
		if posting.IsCleared {
			b.WriteString("* ")
		}
		b.WriteString(posting.Account)
		b.WriteString(p.postingAccountPostSpace(x, posting))
		if posting.BalanceAssertion != nil {
			b.WriteString("= ")
			b.WriteString(amount(posting.BalanceAssertion))
		}
		if posting.Amount != nil {
			b.WriteString(amount(posting.Amount))
		}
		if posting.LotPrice != nil {
			b.WriteString(" { ")
			b.WriteString(amount(posting.LotPrice))
			b.WriteString(" }")
		}
		if !posting.LotDate.IsZero() {
			b.WriteString(" [")
			b.WriteString(toDate(posting.LotDate))
			b.WriteByte(']')
		}
		if posting.Price != nil {
			b.WriteByte(' ')
			if posting.PriceIsForWhole {
				b.WriteByte('@')
			}
			b.WriteString("@ ")
			b.WriteString(amount(posting.Price))
		}
		if posting.BalanceAssertion != nil {
			b.WriteString(" = ")
			b.WriteString(amount(posting.BalanceAssertion))
		}
		if posting.Note != "" {
			b.WriteString(posting.NotePreSpace)
			b.WriteString(p.commentReturns(x, posting.Note))
		}
	}
	b.WriteByte('\n')
}
