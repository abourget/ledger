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
			`
2007-09-09 Hello ; World
  A  20$
  B  ; Hello world
`,
			`
2007-09-09 Hello ; World
    A                  20$
    B
`,
		},
	}

	for _, test := range tests {
		tree := parse.New("filename", test.in)
		assert.NoError(t, tree.Parse())
		buf := &bytes.Buffer{}
		assert.NoError(t, New(tree).Print(buf))
		assert.Equal(t, test.out, buf.String())
	}
}
