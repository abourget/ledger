package journal

import "math/big"

type Amounts struct {
	ByCommodity map[string]*big.Rat
}

func NewAmounts() *Amounts {
	return &Amounts{
		ByCommodity: make(map[string]*big.Rat),
	}
}

func (a *Amounts) Add(commodity string, quantity string) {
	v, ok := a.ByCommodity[commodity]
	if !ok {
		v = big.NewRat(0, 1)
		a.ByCommodity[commodity] = v
	}
	change, ok := big.NewRat(0, 1).SetString(quantity)
	if !ok {
		panic("Cannot parse quantity: " + quantity)
	}
	v.Add(v, change)
}

func (a *Amounts) Get(commodity string) *big.Rat {
	if v, ok := a.ByCommodity[commodity]; ok {
		return v
	}
	return big.NewRat(0, 1)
}
