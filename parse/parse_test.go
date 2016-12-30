package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tree := New("file.ledger", `

2016/09/09 * Kentucky Friends Chicks Ends  ; Never go back there
  Expenses:Restaurants    20.00 CAD
  Assets:Cash             CAD -20.00

2016/09/09 ! Payee
  Expenses:Misc    20.00 CAD
  Assets:Cash  ; Woah, not sure
`)
	err := tree.Parse()
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, tree.Root.Nodes[0].Type(), NodeXact)
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input string
		error string
	}{
		{`2016/09/09 * * heya!`, "1: cannot specify cleared and/or pending more than once"},
	}

	for _, test := range tests {
		tree := New("file.ledger", test.input)
		err := tree.Parse()
		assert.Error(t, err)
		assert.Equal(t, "ledger: file.ledger:"+test.error, err.Error())
	}
}
