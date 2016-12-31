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
		it := t.next()
		switch it.typ {
		case itemError:
			t.errorf(it.val)
		case itemSpace, itemEOL:
			spaceVal := it.val
		EatSpaces:
			for {
				switch it := t.peek(); it.typ {
				case itemSpace, itemEOL:
					t.next()
					spaceVal += it.val
				default:
					break EatSpaces
				}
			}
			t.Root.add(t.newSpace(it.pos, spaceVal))
		case itemComment:
			t.Root.add(t.newComment(it))
			t.expect(itemEOL, "comment")
		case itemEqual:
			// Analyze an automated transaction
		case itemTilde:
			// Analyze a periodic transaction
		case itemDate:
			// Analyze a plain transaction
			txDate, err := parseDate(it.val)
			if err != nil {
				t.error(err)
			}
			x := t.newXact(it.pos)
			x.Date = txDate
			t.parseXact(x)
		default:
			t.errorf("unsupported top-level directive")
		}
	}
	return nil
}

func (t *Tree) parseXact(x *XactNode) {
	switch it := t.peekNonSpace(); it.typ {
	case itemEqual:
		t.next() // consume this one
		it = t.nextNonSpace()
		if it.typ != itemDate {
			t.unexpected(it, "transaction, after '='")
		}
		effectiveDate, err := parseDate(it.val)
		if err != nil {
			t.error(err)
		}
		x.EffectiveDate = effectiveDate
	}

	switch it := t.peekNonSpace(); it.typ {
	case itemAsterisk:
		t.next()
		x.IsCleared = true
	case itemExclamation:
		t.next()
		x.IsPending = true
	}

	switch it := t.peekNonSpace(); it.typ {
	case itemLeftParen:
		t.next()
		if it := t.next(); it.typ != itemString {
			t.unexpected(it, "transaction code, expected a string")
		} else {
			x.Code = it.val
			t.expect(itemRightParen, "transaction code closing")
		}
	}

	switch it := t.peekNonSpace(); it.typ {
	case itemString:
		t.next()
		x.Description = it.val
	case itemAsterisk, itemExclamation:
		t.errorf("cannot specify cleared and/or pending more than once")
	case itemNote:
		t.errorf("missing payee/description before notes")
	case itemEOL, itemEOF:
		t.errorf("unexpected end of input")
	}

	switch it := t.peekNonSpace(); it.typ {
	case itemNote:
		t.next()
		x.Note = it.val
	default:
	}

	t.expect(itemEOL, "transaction opening line")

	if x.Note == "" {
		x.Description = strings.TrimRight(x.Description, " ")
	}

	t.parsePostings(x)
}

func (t *Tree) parsePostings(x *XactNode) {
	// stop on double EOL, or EOL + Space + EOL
	var posting *PostingNode
	for {
		switch it := t.peek(); it.typ {
		case itemSpace:
			t.next()
			// This is posting
			switch it := t.peek(); it.typ {
			case itemAccountName:
				t.next()
				posting = x.newPosting(it.pos)
				posting.Account = it.val
				t.parsePosting(posting)

			case itemNote:
				t.next()
				if posting == nil {
					// attach to the XactMode
					if x.Note == "" {
						x.Note = it.val
					} else {
						x.Note = x.Note + "\n" + it.val
					}
				} else {
					posting.Note = posting.Note + "\n" + it.val
				}
				t.expect(itemEOL, "comment")
			}

		default:
			return
		}
	}
}

func (t *Tree) parsePosting(p *PostingNode) {
	switch it := t.peek(); it.typ {
	case itemSpace:
		t.next()
		p.AccountPostSpace = it.val
	case itemEOL:
		t.next()
		return
	default:
		t.unexpected(it, "transaction posting")
		return
	}

	if it := t.peek(); it.typ == itemNote {
		t.next()
		p.Note = it.val

		t.expect(itemEOL, "transaction opening line")
		return
	}

	/**
	amount:
	    neg_opt commodity neg_opt quantity annotation |
	    neg_opt quantity commodity annotation ;
	*/

	a := t.parseAmount()
	p.Amount = a

	// Parse optional prices '@' and '@@'
	if it := t.peekNonSpace(); it.typ == itemAt || it.typ == itemDoubleAt {
		t.next()

		if it.typ == itemDoubleAt {
			p.PriceIsForWhole = true
		}

		a := t.parseAmount()
		if a.Negative {
			t.unexpected(it, "posting price; a negative price ?")
		}
		p.Price = a
	}

	// Handle "= AMOUNT"
	if it := t.peekNonSpace(); it.typ == itemEqual {
		t.next()
		a := t.parseAmount()
		if p.Amount == nil {
			p.BalanceAssignment = a
		} else {
			p.BalanceAssertion = a
		}
	}

	if it := t.peekNonSpace(); it.typ == itemLotPrice {
		t.next()
		p.LotPrice = t.newAmount()
		p.LotPrice.Quantity = strings.Trim(it.val, "{}")
		p.LotPrice.Pos = it.pos
	}

	if it := t.peekNonSpace(); it.typ == itemLotDate {
		t.next()

		dt, err := parseDate(it.val)
		if err != nil {
			t.error(err)
		}
		p.LotDate = dt
	}

	if it := t.peek(); it.typ == itemSpace {
		t.next()
		p.NotePreSpace = it.val
	}

	if it := t.peek(); it.typ == itemNote {
		t.next()
		p.Note = it.val
	}

	t.expect(itemEOL, "transaction opening line")
}

func (t *Tree) parseAmount() (amount *AmountNode) {
	amount = t.newAmount()

	// TODO: implement detection of value expressions
	// https://github.com/ledger/ledger/blob/next/doc/grammar.y#L149-L155

	switch it := t.peek(); it.typ {
	case itemNeg:
		amount.next(t)
		amount.Negative = true
	case itemEqual:
		return nil
	}

	amount.space(t)

	switch it := t.peek(); it.typ {
	case itemCommodity:
		amount.next(t)
		amount.Commodity = it.val
	case itemQuantity:
		amount.next(t)
		amount.Quantity = it.val

	default:
		t.unexpected(it, "amount quantity/commodity")
	}

	amount.space(t)

	if it := t.peek(); it.typ == itemNeg {
		if amount.Negative {
			t.errorf("unexpected double negative")
		}
		amount.next(t)
		amount.Negative = true
	}

	amount.space(t)

	switch it := t.peek(); it.typ {
	case itemCommodity:
		if amount.Commodity != "" {
			t.errorf("unexpected commodity (specified twice ?)")
		}
		amount.next(t)
		amount.Commodity = it.val
	case itemQuantity:
		if amount.Quantity != "" {
			t.errorf("unexpected quantity (specified twice ?)")
		}
		amount.next(t)
		amount.Quantity = it.val
	default:
		t.unexpected(it, "amount")
	}

	return
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

func appendComment(orig, new string) string {
	if orig == "" {
		return new
	}
	return orig + "\n" + new
}

func parseDate(input string) (time.Time, error) {
	stdSeparator := strings.Replace(strings.Replace(input, "/", "-", -1), ".", "-", -1)
	undecorated := strings.Trim(stdSeparator, "[]") // from itemLotPrice
	return time.ParseInLocation("2006-01-02", undecorated, time.UTC)
}
