package journal

import "math/big"

type Account struct {
	Name    string
	Amounts map[string]*Amount
}

func (a Account) String() string {
	s := ""
	for _, am := range a.Amounts {
		if s != "" {
			s += "\n"
		}
		s += am.String()
	}
	return s + "  " + a.Name
}

func (a *Account) Add(am *Amount) {
	accAmount, ok := a.Amounts[am.Commodity]
	if !ok {
		accAmount = &Amount{
			Commodity: am.Commodity,
			Quantity:  big.NewRat(0, 1),
		}
		a.Amounts[am.Commodity] = accAmount
	}
	accAmount.Quantity.Add(accAmount.Quantity, am.Quantity)
}

func NewAccount(name string) *Account {
	return &Account{
		Name:    name,
		Amounts: make(map[string]*Amount),
	}
}
