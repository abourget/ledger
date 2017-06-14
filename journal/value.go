package journal

import (
	"fmt"
	"math/big"
	"strings"
)

type Amount struct {
	Commodity string
	Quantity  *big.Rat
}

func amountToString(v interface{}) string {
	switch vv := v.(type) {
	case string:
		return vv
	case int:
		return fmt.Sprintf("%d", vv)
	case int64:
		return fmt.Sprintf("%d", vv)
	case float32:
		return fmt.Sprintf("%g", vv)
	case float64:
		return fmt.Sprintf("%g", vv)
	case *big.Rat:
		return vv.String()
	default:
		return ""
	}
}

func (a Amount) String() string {
	q := strings.TrimRight(a.Quantity.FloatString(10), "0")
	q = strings.TrimRight(q, ".")
	return q + " " + a.Commodity
}
