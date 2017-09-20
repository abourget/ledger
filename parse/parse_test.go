package parse

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tree := New("file.ledger", `
; Top level comment

2016/09/09 = 2016-09-10 * Kentucky Friends Beef      ; Never go back there
   ; Some more notes for the transaction
  Expenses:Restaurants    20.00 CAD
  Assets:Cash             CAD -20.00    ; That hurt


2016/09/09 ! (Kode) Payee
  ; Transaction notes
  Expenses:Misc    20.00 CAD = 700.00 CAD
  Expenses:Tip     50.00 USD @ 100.00 CAD
  Assets:Cash  ; Woah, not sure
 ; Here again
  ; And yet another note for this posting.

2016/09/10 * Hi there
  A   = 23 CAD
  B   -100 CAD = -200 USD
`)
	err := tree.Parse()
	require.NoError(t, err)
	assert.Len(t, tree.Root.Nodes, 8)

	spc, ok := tree.Root.Nodes[0].(*SpaceNode)
	require.True(t, ok)
	assert.Equal(t, "\n", spc.Space)

	comm, ok := tree.Root.Nodes[1].(*CommentNode)
	require.True(t, ok)
	assert.Equal(t, "; Top level comment", comm.Comment)

	spc, ok = tree.Root.Nodes[2].(*SpaceNode)
	require.True(t, ok)
	assert.Equal(t, "\n", spc.Space)

	// Transaction 1
	xact, ok := tree.Root.Nodes[3].(*XactNode)
	require.True(t, ok)
	assert.Equal(t, "Kentucky Friends Beef      ", xact.Description)
	assert.Equal(t, "; Never go back there\n; Some more notes for the transaction", xact.Note)
	assert.Equal(t, "  ", xact.Postings[0].AccountPreSpace)
	assert.Equal(t, "Expenses:Restaurants", xact.Postings[0].Account)
	assert.Equal(t, "    ", xact.Postings[0].AccountPostSpace)
	assert.Equal(t, "20.00 CAD", xact.Postings[0].Amount.Raw)
	assert.Equal(t, "20.00", xact.Postings[0].Amount.Quantity)
	assert.Equal(t, "CAD", xact.Postings[0].Amount.Commodity)
	assert.True(t, xact.IsCleared)
	assert.False(t, xact.IsPending)
	assert.Equal(t, xact.EffectiveDate, time.Date(2016, time.September, 10, 0, 0, 0, 0, time.UTC))

	// Spacing
	spaces, ok := tree.Root.Nodes[4].(*SpaceNode)
	require.True(t, ok)
	assert.Equal(t, "\n\n", spaces.Space)

	// Transaction 2
	xact, ok = tree.Root.Nodes[5].(*XactNode)
	require.True(t, ok)
	assert.Equal(t, "Expenses:Tip", xact.Postings[1].Account)
	assert.Equal(t, "50.00 USD", xact.Postings[1].Amount.Raw)
	assert.Equal(t, " 100.00 CAD", xact.Postings[1].Price.Raw)
	assert.Equal(t, "100.00", xact.Postings[1].Price.Quantity)

	assert.False(t, xact.IsCleared)
	assert.True(t, xact.IsPending)

	//treeToJSON(tree)
}

func TestParseEdgeCases(t *testing.T) {
	tree := New("file.ledger", `
2016/09/10 Desc
  A  - $ 23
  B  23 $ @@ 2 CAD

2016/09/10 Desc 2
  ! A      (23 CAD + 123 USD)
  ! B
2016/10/10 Desc 3
  A  $-12
  B  $.34
  CWithTrailingSpaces     
`)
	err := tree.Parse()
	require.NoError(t, err)
	assert.Len(t, tree.Root.Nodes, 5)

	spaces, ok := tree.Root.Nodes[0].(*SpaceNode)
	require.True(t, ok)
	assert.Equal(t, "\n", spaces.Space)

	xact, ok := tree.Root.Nodes[1].(*XactNode)
	require.True(t, ok)

	assert.Equal(t, "Desc", xact.Description)
	assert.Equal(t, "A", xact.Postings[0].Account)
	assert.Equal(t, "- $ 23", xact.Postings[0].Amount.Raw)
	assert.Equal(t, true, xact.Postings[0].Amount.Negative)
	assert.Equal(t, "$", xact.Postings[0].Amount.Commodity)
	assert.Equal(t, "B", xact.Postings[1].Account)
	assert.Equal(t, "23 $", xact.Postings[1].Amount.Raw)
	assert.Equal(t, "$", xact.Postings[1].Amount.Commodity)
	assert.Equal(t, "23", xact.Postings[1].Amount.Quantity)
	assert.Equal(t, " 2 CAD", xact.Postings[1].Price.Raw)
	assert.Equal(t, "CAD", xact.Postings[1].Price.Commodity)
	assert.Equal(t, "2", xact.Postings[1].Price.Quantity)
	assert.Equal(t, true, xact.Postings[1].PriceIsForWhole)

	xact, ok = tree.Root.Nodes[3].(*XactNode)
	require.True(t, ok)

	assert.Equal(t, "Desc 2", xact.Description)
	assert.Equal(t, "A", xact.Postings[0].Account)
	assert.Equal(t, "(23 CAD + 123 USD)", xact.Postings[0].Amount.Raw)
	assert.Equal(t, "(23 CAD + 123 USD)", xact.Postings[0].Amount.ValueExpr)
	assert.Equal(t, false, xact.Postings[0].Amount.Negative)
	assert.Equal(t, true, xact.Postings[0].IsPending)
	assert.Equal(t, false, xact.Postings[0].IsCleared)
	assert.Equal(t, "B", xact.Postings[1].Account)
	assert.Nil(t, xact.Postings[1].Amount)
	assert.Nil(t, xact.Postings[1].Price)

	xact, ok = tree.Root.Nodes[4].(*XactNode)
	require.True(t, ok)

	assert.Equal(t, "Desc 3", xact.Description)
	assert.Equal(t, "A", xact.Postings[0].Account)
	assert.Equal(t, "$-12", xact.Postings[0].Amount.Raw)
	assert.Equal(t, true, xact.Postings[0].Amount.Negative)
	assert.Equal(t, "12", xact.Postings[0].Amount.Quantity)
	assert.Equal(t, "$", xact.Postings[0].Amount.Commodity)
	assert.Equal(t, "B", xact.Postings[1].Account)
	assert.Equal(t, "$.34", xact.Postings[1].Amount.Raw)
	assert.Equal(t, false, xact.Postings[1].Amount.Negative)
	assert.Equal(t, ".34", xact.Postings[1].Amount.Quantity)
	assert.Equal(t, "$", xact.Postings[1].Amount.Commodity)

	treeToJSON(tree)
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

func treeToJSON(t *Tree) {
	a, _ := json.MarshalIndent(t.Root, "", "  ")
	os.Stdout.Write(a)
}
