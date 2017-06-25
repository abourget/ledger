package budget

import (
	"time"

	"github.com/xconstruct/ledger/journal"
	"github.com/xconstruct/ledger/tools/filter"
	"github.com/xconstruct/ledger/tools/reports"
)

func FindBudgetTxs(txs []*journal.Transaction) []*journal.Transaction {
	return filter.New(txs, filter.Note("budget:")).Slice()
}

func Balance(j *journal.Journal, since time.Time) (*reports.BalanceReport, error) {
	txs, err := j.Transactions()
	if err != nil {
		return nil, err
	}

	accounts := make(map[string]bool)
	budget := FindBudgetTxs(txs)
	for _, b := range budget {
		bdate := b.Node.Date
		for _, p := range b.Postings() {
			accounts[p.Account()] = true
		}

		now := time.Now()
		sy, sm, sd := bdate.Date()
		ey, em, _ := now.Date()
		months := time.Month(ey-sy)*12 + em - sm
		for m := time.Month(0); m <= months; m++ {
			date := time.Date(sy, sm+m, sd, 4, 0, 0, 0, time.Local)
			if bdate.After(date) || date.After(now) {
				continue
			}

			tx := j.AddTransaction(date, "budget")
			for _, p := range b.Postings() {
				a := p.Amount()
				tx.NewPosting(p.Account()).SetAmount(a.Commodity, a.Quantity)
			}
		}
	}

	txs, err = j.Transactions()
	txs = filter.New(txs, filter.Since(since), filter.Not(filter.Note("budget:"))).Slice()
	bal := reports.Balance(txs)
	bal2 := &reports.BalanceReport{
		Accounts: make(map[string]*reports.Account),
	}
	for acc := range accounts {
		if a, ok := bal.Accounts[acc]; ok {
			bal2.Accounts[acc] = a
		}
	}
	return bal2, err
}
