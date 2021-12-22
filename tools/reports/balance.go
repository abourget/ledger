// Package reports provides simple aggregation and reporting of ledger files.
package reports

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/abourget/ledger/journal"
	"github.com/abourget/ledger/lpath"
)

type BalanceReport struct {
	Accounts map[string]*journal.Account
	Total    *journal.Account
}

func (b *BalanceReport) account(name string) *journal.Account {
	if name == "" {
		return b.Total
	}

	acc, ok := b.Accounts[name]
	if !ok {
		acc = journal.NewAccount(name)
		b.Accounts[name] = acc
	}
	return acc
}

func Balance(txs []*journal.Transaction) *BalanceReport {
	return BalanceFiltered(txs, nil)
}

func relevant(f func(account string) bool, account string) bool {
	for account != "" {
		if f(account) {
			return true
		}
		account = lpath.Base(account)
	}
	return false
}

func BalanceFiltered(txs []*journal.Transaction, filter func(account string) bool) *BalanceReport {
	b := &BalanceReport{
		Accounts: make(map[string]*journal.Account),
		Total:    journal.NewAccount(""),
	}

	// Sum up balances of all individual accounts
	for _, tx := range txs {
		for _, p := range tx.Postings() {
			b.account(p.Account()).Add(p.Amount())
		}
	}

	// Filter accounts
	// relevant: accounts that influence the balance (e.g. hidden child accounts)
	// displayed: accounts that should be displayed (e.g. irrelevant parent accounts)
	displayed := make(map[string]bool)
	for _, acc := range b.Accounts {

		relevant := false
		name := acc.Name
		for name != "" {
			if filter == nil || filter(name) {
				relevant = true
			}
			if relevant {
				displayed[name] = true
			}
			name = lpath.Base(name)
		}

		if !relevant {
			delete(b.Accounts, acc.Name)
		}
	}

	// Make sure all parent accounts are created
	for name := range displayed {
		b.account(name)
	}

	// Sort by granularity (children before parents)
	accs := make([]*journal.Account, 0)
	for _, acc := range b.Accounts {
		accs = append(accs, acc)
	}
	sort.Slice(accs, func(i, j int) bool {
		a, b := accs[i].Name, accs[j].Name
		an, bn := strings.Count(a, ":"), strings.Count(b, ":")
		if an == bn {
			return a < b
		}
		return an > bn
	})
	// Aggregate child accounts into parents
	for _, acc := range accs {
		base := lpath.Base(acc.Name)
		for _, am := range acc.Amounts {
			b.account(base).Add(am)
		}
	}

	// Delete hidden child accounts
	for name := range b.Accounts {
		if !displayed[name] {
			delete(b.Accounts, name)
		}
	}

	return b
}

func (b *BalanceReport) Print(w io.Writer) error {
	list := make([]*formattedAccount, 0, len(b.Accounts))
	tf := formatAccount(b.Total)
	length := tf.Length
	for _, acc := range b.Accounts {
		f := formatAccount(acc)
		list = append(list, f)
		if f.Length > length {
			length = f.Length
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	for _, f := range list {
		_, err := fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", length-f.Length), f.String)
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "%s\n%s%s\n", strings.Repeat("-", length),
		strings.Repeat(" ", length-tf.Length), tf.String)
	if err != nil {
		return err
	}

	return nil
}
