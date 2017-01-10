package print

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
    tests := []struct{
        in string
        out string
    }{
        {
			`
2007-09-09 Hello ; World
  A  20$
  B
`,
			`
2007-09-09 Hello ; World
    A                  20$
    B
`
		},
    }

    for _, test := range tests {
		assert.NoError(parse.New(test.in).Parse())
        res := New((test.in)
        assert.Equal(t, test.out, res)
    }
}
