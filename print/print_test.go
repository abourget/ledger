package print

import (
	"bytes"
	"testing"

	"github.com/abourget/ledger/parse"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			`; comment
; second
2016/01/01=2016.02/02 Tx ; another comment
  Account1:Hello World     10.00$    @   12.23 USD  ; Note 7 flames
  Other                    (123 USD)  ; Note

2016/01/01 !Tx
  Account1:Hello World          $10.00 [2017/01/01]  ; Then comment
  ! Other  ; Comment here
  ; Comment there

2017/1/1 * (kode) Tx
 Account1:Hello World        - 10.00 $
 Other                   (10.00 $ * 2)
`,
			`; comment
; second
2016-01-01 = 2016-02-02 Tx ; another comment
    Account1:Hello World              $10.00 @ 12.23 USD  ; Note 7 flames
    Other                             (123 USD)  ; Note

2016-01-01 ! Tx
    Account1:Hello World              $10.00 [2017-01-01]  ; Then comment
    ! Other                           ; Comment here
    ; Comment there

2017-01-01 * (kode) Tx
    Account1:Hello World              -$10.00
    Other                             (10.00 $ * 2)
`,
		},
	}

	for _, test := range tests {
		tree := parse.New("filename", test.in)
		assert.NoError(t, tree.Parse())
		buf := &bytes.Buffer{}
		printer := New(tree)
		printer.MinimumAccountWidth = 30
		assert.NoError(t, printer.Print(buf))
		assert.Equal(t, test.out, buf.String())
	}
}
