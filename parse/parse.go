package parse

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
	"time"
)

// Tree is the representation of a single parsed Ledger file.
type Tree struct {
	FileName  string    // name of the template represented by the tree.
	Root      *ListNode // top-level root of the tree.
	text      string    // text parsed to create the template (or its parent)
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

func Parse(filename string) (t *Tree, err error) {
	cnt, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	t = New(filename, string(cnt))

	return t, t.Parse()
}

// New creates a new Parse Tree for a Ledger file.  Call .Parse() with
// parameter to retrieve an AST and parsed nodes.
func New(name, input string) *Tree {
	return &Tree{
		FileName: name,
		text:     input,
		lex:      lex(name, input),
	}
}

// parse is the top-level parser for a Ledger file. It runs to EOF.
func (t *Tree) Parse() (err error) {
	defer t.recover(&err)

	t.Root = t.newList(Pos(0))

	for t.peek().typ != itemEOF {
		token := t.next()
		switch token.typ {
		case itemError:
			t.errorf(token.val)
		case itemSpace:
			t.Root.add(t.newSpace(token))
		case itemComment:
			t.Root.add(t.newComment(token))
		case itemEqual:
			// Analyze an automated transaction
		case itemTilde:
			// Analyze a periodic transaction
		case itemDate:
			// Analyze a plain transaction
			txDate, err := time.Parse("2006-01-02", strings.Replace(strings.Replace(token.val, "/", "-", -1), ".", "-", -1))
			if err != nil {
				t.error(err)
			}
			x := t.newXact(token.pos)
			x.add(token)
			x.Date = txDate
			t.parseXact(x)
		}
	}
	return nil
}

func (t *Tree) parseXact(x *XactNode) {
	switch n := t.peekNonSpace(); n.typ {
	case itemAsterisk:
		t.next()
		x.IsCleared = true
	case itemExclamation:
		t.next()
		x.IsPending = true
	case itemNote:
		t.errorf("missing payee/description before notes")
	case itemEOL, itemEOF:
		t.errorf("unexpected end of input")
	}

	switch n := t.peekNonSpace(); n.typ {
	case itemAsterisk, itemExclamation:
		t.errorf("cannot specify cleared and/or pending more than once")
	case itemString:
		token := t.next()
		x.Description = token.val
	case itemNote:
		t.errorf("missing payee/description before notes")
	case itemEOL, itemEOF:
		t.errorf("unexpected end of input")
	}

}

func (t *Tree) append(n Node) {
	t.Root.Nodes = append(t.Root.Nodes, n)
}

// next returns the next token.
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (t *Tree) backup3(t2, t1 item) { // Reverse order: we're pushing back.
	t.token[1] = t1
	t.token[2] = t2
	t.peekCount = 3
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// nextNonSpace returns the next non-space token.
func (t *Tree) nextNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSpace {
			break
		}
	}
	return token
}

// peekNonSpace returns but does not consume the next non-space token.
func (t *Tree) peekNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSpace {
			break
		}
	}
	t.backup()
	return token
}

// ErrorContext returns a textual representation of the location of the node in the input text.
// The receiver is only used when the node does not have a pointer to the tree inside,
// which can occur in old code.
func (t *Tree) ErrorContext(n Node) (location, context string) {
	pos := int(n.Position())
	tree := n.tree()
	if tree == nil {
		tree = t
	}
	text := tree.text[:pos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	lineNum := 1 + strings.Count(text, "\n")
	context = n.String()
	if len(context) > 20 {
		context = fmt.Sprintf("%.20s...", context)
	}
	return fmt.Sprintf("%s:%d:%d", tree.FileName, lineNum, byteNum), context
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	format = fmt.Sprintf("ledger: %s:%d: %s", t.FileName, t.lex.lineNumber(), format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing.
func (t *Tree) error(err error) {
	t.errorf("%s", err)
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected itemType, context string) item {
	token := t.nextNonSpace()
	if token.typ != expected {
		t.unexpected(token, context)
	}
	return token
}

// expectOneOf consumes the next token and guarantees it has one of the required types.
func (t *Tree) expectOneOf(expected1, expected2 itemType, context string) item {
	token := t.nextNonSpace()
	if token.typ != expected1 && token.typ != expected2 {
		t.unexpected(token, context)
	}
	return token
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token item, context string) {
	t.errorf("unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.lex.drain()
			//t.stopParse()
		}
		*errp = e.(error)
	}
	return
}
