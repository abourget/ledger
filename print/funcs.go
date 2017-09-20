package print

import (
	"strings"
	"text/template"
	"time"

	"github.com/abourget/ledger/parse"
)

func funcsPlainXact(minimumAccountWidth, prefixWidth int) template.FuncMap {
	return template.FuncMap{
		"posting_account_pre_space": func(node *parse.XactNode, post *parse.PostingNode) string {
			if prefixWidth == 0 {
				return node.Postings[0].AccountPreSpace
			}
			return strings.Repeat(" ", prefixWidth)
		},
		"posting_account_post_space": func(node *parse.XactNode, post *parse.PostingNode) string {
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

			if longestAccountName < minimumAccountWidth {
				longestAccountName = minimumAccountWidth
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
		},
		"amount": func(amount *parse.AmountNode) (out string) {
			if amount.ValueExpr != "" {
				return amount.ValueExpr
			}
			if amount.Negative {
				out += "-"
			}
			out += amount.Quantity
			if amount.Commodity != "" {
				out += " " + amount.Commodity
			}
			return out
		},
		"to_date": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"comment_returns": func(node *parse.XactNode, input string) string {
			width := prefixWidth
			if width == 0 {
				width = len(node.Postings[0].AccountPreSpace)
			}
			return strings.Replace(input, "\n", "\n"+strings.Repeat(" ", width), -1)
		},
	}
}

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
