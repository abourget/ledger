package util

import (
	"math/big"
	"sort"
	"strings"

	"github.com/xconstruct/ledger/journal"
)

type BalanceReport struct {
	Accounts map[string]*Account
}

func (b *BalanceReport) account(name string) *Account {
	acc, ok := b.Accounts[name]
	if !ok {
		acc = &Account{name, make(map[string]*journal.Amount)}
		b.Accounts[name] = acc
	}
	return acc
}

type Account struct {
	Name    string
	Amounts map[string]*journal.Amount
}

func (a *Account) Add(am *journal.Amount) {
	accAmount, ok := a.Amounts[am.Commodity]
	if !ok {
		accAmount = &journal.Amount{am.Commodity, big.NewRat(0, 1)}
		a.Amounts[am.Commodity] = accAmount
	}
	accAmount.Quantity.Add(accAmount.Quantity, am.Quantity)
}

func Balance(txs []*journal.Transaction) *BalanceReport {
	b := &BalanceReport{
		Accounts: make(map[string]*Account),
	}

	for _, tx := range txs {
		for _, p := range tx.Postings() {
			b.account(p.Account()).Add(p.Amount())
		}
	}

	accs := make([]*Account, 0)
	for _, acc := range b.Accounts {
		accs = append(accs, acc)
	}
	sort.Slice(accs, func(i, j int) bool {
		a, b := accs[i].Name, accs[j].Name
		if strings.Count(a, ":") > strings.Count(b, ":") {
			return true
		}
		return a < b
	})
	for _, acc := range accs {
		if base := journal.BaseAccount(acc.Name); base != "" {
			for _, am := range acc.Amounts {
				b.account(base).Add(am)
			}
		}
	}

	return b
}
