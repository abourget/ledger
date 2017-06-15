package parse

import (
	"fmt"
	"testing"
)

// Make the types prettyprint.
var itemName = map[itemType]string{
	itemError: "error",
	itemDot:   ".",
	itemSpace: "spaces",
	itemText:  "text",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

type lexTest struct {
	name  string
	input string
	items []item
}

var (
	tEOF = item{itemEOF, 0, ""}
	tEOL = item{itemEOL, 0, "\n"}
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{{itemSpace, 0, " \t"}, tEOL, tEOF}},
	{"auto xact", `= `, []item{
		{itemEqual, 0, "="},
		{itemSpace, 0, " "},
		tEOF,
	}},
	{"periodic xact with period", `~  monthly ; Note`, []item{
		{itemTilde, 0, "~"},
		{itemSpace, 0, "  "},
		{itemString, 0, "monthly "},
		{itemNote, 0, "; Note"},
		tEOF,
	}},
	{"plain xact", "2016/09/09 Payee", []item{
		{itemDate, 0, "2016/09/09"},
		{itemSpace, 0, " "},
		{itemString, 0, "Payee"},
		tEOF,
	}},
	{"plain xact eof with note", "2016/09/08 Payee", []item{
		{itemDate, 0, "2016/09/08"},
		{itemSpace, 0, " "},
		{itemString, 0, "Payee"},
		tEOF,
	}},
	{"include file", `include "filename"`, []item{
		{itemInclude, 0, "include"},
		{itemSpace, 0, " "},
		{itemString, 0, `"filename"`},
		tEOF,
	}},
	{"periodic xact truncated", `~ `, []item{
		{itemTilde, 0, "~"},
		{itemSpace, 0, " "},
		tEOF,
	}},
	{"periodic xact missing period", `~  ; Note`, []item{
		{itemTilde, 0, "~"},
		{itemSpace, 0, "  "},
		{itemNote, 0, "; Note"},
		tEOF,
	}},

	{"simple transaction", "2016/09/09 Payee\n Account  - 20.00 CAD", []item{
		{itemDate, 0, "2016/09/09"},
		{itemSpace, 0, " "},
		{itemString, 0, "Payee"},
		tEOL,
		{itemSpace, 0, " "},
		{itemAccountName, 0, "Account"},
		{itemSpace, 0, "  "},
		{itemNeg, 0, "-"},
		{itemSpace, 0, " "},
		{itemQuantity, 0, "20.00"},
		{itemSpace, 0, " "},
		{itemCommodity, 0, "CAD"},
		tEOF,
	}},
	{"less simple transaction", "2016/09/09 * Payee ; So help me God\n    Account  -20.00 CAD\n    Account2:Spaced child:Leaf     CAD 20.00\n", []item{
		{itemDate, 0, "2016/09/09"},
		{itemSpace, 0, " "},
		{itemAsterisk, 0, "*"},
		{itemSpace, 0, " "},
		{itemString, 0, "Payee "},
		{itemNote, 0, "; So help me God"},
		tEOL,
		{itemSpace, 0, "    "},
		{itemAccountName, 0, "Account"},
		{itemSpace, 0, "  "},
		{itemNeg, 0, "-"},
		{itemQuantity, 0, "20.00"},
		{itemSpace, 0, " "},
		{itemCommodity, 0, "CAD"},
		tEOL,
		{itemSpace, 0, "    "},
		{itemAccountName, 0, "Account2:Spaced child:Leaf"},
		{itemSpace, 0, "     "},
		{itemCommodity, 0, "CAD"},
		{itemSpace, 0, " "},
		{itemQuantity, 0, "20.00"},
		tEOL,
		tEOF,
	}},
	{"transaction with price", "2016/09/09 Payee\n Account  - 20.00 CAD @ USD 40.00", []item{
		{itemDate, 0, "2016/09/09"},
		{itemSpace, 0, " "},
		{itemString, 0, "Payee"},
		tEOL,
		{itemSpace, 0, " "},
		{itemAccountName, 0, "Account"},
		{itemSpace, 0, "  "},
		{itemNeg, 0, "-"},
		{itemSpace, 0, " "},
		{itemQuantity, 0, "20.00"},
		{itemSpace, 0, " "},
		{itemCommodity, 0, "CAD"},
		{itemSpace, 0, " "},
		{itemAt, 0, "@"},
		{itemSpace, 0, " "},
		{itemCommodity, 0, "USD"},
		{itemSpace, 0, " "},
		{itemQuantity, 0, "40.00"},
		tEOF,
	}},
	{"price directive", `P 2017/06/15 USD 50.00 CAD`, []item{
		{itemPrice, 0, "P"},
		{itemSpace, 0, " "},
		{itemString, 0, "2017/06/15 USD 50.00 CAD"},
		tEOF,
	}},

	// errors

	{"plain xact eof", "2016/09/09", []item{
		{itemDate, 0, "2016/09/09"},
		{itemError, 0, "unexpected end-of-file"},
	}},
	{"plain xact eof with note", "2016/09/09\n", []item{
		{itemDate, 0, "2016/09/09"},
		{itemError, 0, "unexpected end-of-line"},
	}},
	{"erroneous date non-digit", "2016/09eee\n", []item{
		{itemError, 0, "date format error, expects YYYY-MM-DD with '/', '-' or '.' as separators, received character U+0065 'e'"},
	}},
	{"erroneous date", "2016/099/08 Payee", []item{
		{itemError, 0, "date format error, expects YYYY-MM-DD with '/', '-' or '.' as separators, received character U+0039 '9'"},
	}},
	{"erroneous short date", "2016/09", []item{
		{itemError, 0, "date format error, expects YYYY-MM-DD with '/', '-' or '.' as separators, received character U+FFFFFFFFFFFFFFFF"},
	}},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		// if test.name != "less simple transaction" {
		// 	continue
		// }
		items := collect(&test)
		if !equal(items, test.items, false) {
			t.Errorf("test %q: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := lex(t.name, t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func equal(i1, i2 []item, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
	}
	return true
}
