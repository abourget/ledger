package parse

import (
	"encoding/json"
	"fmt"
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
  Assets:Cash  ; Woah, not sure
 ; Here again
  ; And yet another note for this posting.

2016/09/10 Desc
  A  - $ 23
  B  23 $ @ 2 CAD

2016/09/10 * Hi there
  A   = 23 CAD
  B   -100 CAD = -200 USD
`)
	err := tree.Parse()
	if err != nil {
		t.Error(err.Error())
	}

	fmt.Printf("%#v\n", tree.Root.Nodes)
	assert.Len(t, tree.Root.Nodes, 8)

	a, _ := json.MarshalIndent(tree.Root, "", "  ")
	os.Stdout.Write(a)

	_, ok := tree.Root.Nodes[0].(*SpaceNode)
	require.True(t, ok)

	comm, ok := tree.Root.Nodes[1].(*CommentNode)
	require.True(t, ok)
	assert.Equal(t, "; Top level comment", comm.Comment)

	_, ok = tree.Root.Nodes[2].(*SpaceNode)
	require.True(t, ok)

	// Transaction 1
	xact, ok := tree.Root.Nodes[3].(*XactNode)
	require.True(t, ok)
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

	assert.False(t, xact.IsCleared)
	assert.True(t, xact.IsPending)

	// Spacing
	spaces, ok = tree.Root.Nodes[6].(*SpaceNode)
	require.True(t, ok)
	assert.Equal(t, "\n", spaces.Space)

	// Transaction 3
	xact, ok = tree.Root.Nodes[7].(*XactNode)
	require.True(t, ok)

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
