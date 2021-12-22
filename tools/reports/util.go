package reports

import (
	"strings"

	"github.com/abourget/ledger/journal"
)

type formattedAccount struct {
	Account *journal.Account
	Name    string
	String  string
	Length  int
	Lines   int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatAccount(acc *journal.Account) *formattedAccount {
	f := &formattedAccount{
		Account: acc,
		Name:    acc.Name,
	}
	for _, a := range acc.Amounts {
		v := a.String()
		f.String += v + "\n"
		f.Lines++
		if n := len(v); n > f.Length {
			f.Length = n
		}
	}
	f.String = strings.TrimSpace(f.String) + "  " + f.Name
	return f
}
