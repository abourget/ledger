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
	tDot          = item{itemDot, 0, "."}
	tEOF          = item{itemEOF, 0, ""}
	tEOL          = item{itemEOL, 0, "\n"}
	tSpace        = item{itemSpace, 0, " "}
	tBegAutoXact  = item{itemBeginAutomaticXact, 0, "="}
	tBegPerioXact = item{itemBeginPeriodicXact, 0, "~"}
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{{itemSpace, 0, " \t"}, tEOL, tEOF}},
	{"auto xact", `= `, []item{
		tBegAutoXact,
		{itemError, 0, "not yet implemented"},
	}},
	{"periodic xact with period", `~  monthly ; Note`, []item{
		tBegPerioXact,
		{itemSpace, 0, "  "},
		{itemPeriodExpr, 0, "monthly "},
		{itemNote, 0, "; Note"},
		{itemError, 0, "not yet implemented"},
	}},
	{"plain xact", "2016/09/09 Payee", []item{
		{itemDate, 0, "2016/09/09"},
		{itemSpace, 0, " "},
		{itemError, 0, "not yet implemented"},
	}},
	{"plain xact eof with note", "2016/09---..- Payee", []item{
		{itemDate, 0, "2016/09---..-"},
		{itemSpace, 0, " "},
		{itemError, 0, "not yet implemented"},
	}},
	{"include file", `include "filename"`, []item{
		{itemInclude, 0, "include"},
		{itemSpace, 0, " "},
		{itemString, 0, `"filename"`},
		tEOF,
	}},

	// errors

	{"periodic xact error", `~ `, []item{
		tBegPerioXact,
		{itemSpace, 0, " "},
		{itemError, 0, "premature end-of-file, expected postings for periodic transaction"},
	}},
	{"periodic xact missing period", `~  ; Note`, []item{
		tBegPerioXact,
		{itemSpace, 0, "  "},
		{itemError, 0, "missing period expression"},
	}},
	{"plain xact eof", "2016/09/09", []item{
		{itemError, 0, "unexpected end-of-file, expected transaction Payee or Description"},
	}},
	{"plain xact eof with note", "2016/09\n", []item{
		{itemError, 0, "unexpected end-of-line, expected transaction Payee or Description"},
	}},
	{"plain xact eof with note", "2016/09eee\n", []item{
		{itemError, 0, "invalid character in transaction date specification: 'e'"},
	}},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
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
