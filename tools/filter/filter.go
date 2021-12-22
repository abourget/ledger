// Package filter provides a simple tool to filter ledger transactions
package filter

import (
	"strings"
	"time"

	"github.com/abourget/ledger/journal"
)

type FilterFn func(tx *journal.Transaction) bool

type Filter struct {
	Txs     []*journal.Transaction
	Filters []FilterFn
}

func New(txs []*journal.Transaction, fns ...FilterFn) *Filter {
	return &Filter{
		Txs:     txs,
		Filters: fns,
	}
}

func (f *Filter) Filter(fns ...FilterFn) *Filter {
	f.Filters = append(f.Filters, fns...)
	return f
}

func (f *Filter) Slice() []*journal.Transaction {
	filtered := make([]*journal.Transaction, 0)
TX_LOOP:
	for _, tx := range f.Txs {
		for _, f := range f.Filters {
			if !f(tx) {
				continue TX_LOOP
			}
		}
		filtered = append(filtered, tx)
	}

	return filtered
}

func Since(t time.Time) FilterFn {
	return func(tx *journal.Transaction) bool {
		return tx.Node.Date.After(t)
	}
}

func Until(t time.Time) FilterFn {
	return func(tx *journal.Transaction) bool {
		return t.After(tx.Node.Date)
	}
}

func Account(acc string) FilterFn {
	return func(tx *journal.Transaction) bool {
		return tx.Posting(acc) != nil
	}
}

func Description(desc string) FilterFn {
	return func(tx *journal.Transaction) bool {
		return strings.Contains(tx.Node.Description, desc)
	}
}

func Note(note string) FilterFn {
	return func(tx *journal.Transaction) bool {
		return strings.Contains(" "+tx.Node.Note+" ", " "+note+" ")
	}
}

func Not(f FilterFn) FilterFn {
	return func(tx *journal.Transaction) bool {
		return !f(tx)
	}
}
