package parse

import "strings"

// rawAccumulator fetches items from the lexer, and saves its values,
// as to recreate the original raw form.
type rawAccumulator struct {
	firstPos Pos
	chunks   []string
}

func (a *rawAccumulator) next(t *Tree) item {
	it := t.next()
	if a.firstPos == 0 {
		a.firstPos = it.pos
	}
	a.chunks = append(a.chunks, it.val)
	return it
}

func (a *rawAccumulator) space(t *Tree) {
	if it := t.peek(); it.typ == itemSpace {
		a.next(t)
	}
}

func (a *rawAccumulator) String() string {
	return strings.Join(a.chunks, "")
}
