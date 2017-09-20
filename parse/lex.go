package parse

// TODO: sync with the latest implementation here:
// https://github.com/ledger/ledger4/blob/master/ledger-parse/Ledger/Parser/Text.hs
// hledger: https://github.com/simonmichael/hledger/blob/master/hledger-lib/Hledger/Utils/Parse.hs
// https://github.com/ledger/ledger/blob/next/src/textual.cc as per JohnW, is the end reference.

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	val := fmt.Sprintf("%q", i.val)
	return fmt.Sprintf("%s(%s)", label[i.typ], val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemString
	itemNote    // comments for postings
	itemComment // top-level Journal comments
	itemDate
	itemLotDate
	itemLotPrice
	itemSpace
	itemEOL
	itemText
	itemAt       // "@"
	itemDoubleAt // "@@"
	itemEqual    // '='
	itemAsterisk
	itemExclamation
	itemSemicolon
	itemCommodity
	itemIdentifier
	itemLeftParen
	itemRightParen
	itemNeg       // '-'
	itemQuantity  // "123.1234", with optional decimals. No scientific notation, complex, imaginary, etc..
	itemValueExpr // "(123 + 234)"
	itemTilde
	itemPeriodExpr
	itemDot // to form numbers, with itemInteger + optionally: itemDot + itemInteger
	itemStatus
	itemAccountName // only a name like "Expenses:Misc"
	itemAccount     // can be "(Expenses:Misc)" or "[Expenses:Misc]"
	itemBeginAutomatedXact
	itemBeginPeriodicXact
	itemBeginXact

	// Keywords appear after all the rest.
	itemKeyword
	itemInclude
	itemAccountKeyword
	itemEnd
	itemAlias
	itemPrice
	// itemDef
	// itemYear
	// itemBucket
	// itemAssert
	// itemCheck
	// itemCommodityConversion
	// itemDefaultCommodity
)

// key must contain anything after `itemKeyword` in the preceding list.
var key = map[string]itemType{
	"include": itemInclude,
	"account": itemAccountKeyword,
	"end":     itemEnd,
	"alias":   itemAlias,
	"P":       itemPrice,
}

var label = map[itemType]string{
	itemError:              "itemError",
	itemEOF:                "itemEOF",
	itemString:             "itemString",
	itemNote:               "itemNote",
	itemComment:            "itemComment",
	itemDate:               "itemDate",
	itemLotDate:            "itemLotDate",
	itemLotPrice:           "itemLotPrice",
	itemSpace:              "itemSpace",
	itemEOL:                "itemEOL",
	itemText:               "itemText",
	itemAt:                 "itemAt",
	itemDoubleAt:           "itemDoubleAt",
	itemEqual:              "itemEqual",
	itemAsterisk:           "itemAsterisk",
	itemExclamation:        "itemExclamation",
	itemSemicolon:          "itemSemicolon",
	itemCommodity:          "itemCommodity",
	itemIdentifier:         "itemIdentifier",
	itemLeftParen:          "itemLeftParen",
	itemRightParen:         "itemRightParen",
	itemNeg:                "itemNeg",
	itemQuantity:           "itemQuantity",
	itemValueExpr:          "itemValueExpr",
	itemTilde:              "itemTilde",
	itemPeriodExpr:         "itemPeriodExpr",
	itemDot:                "itemDot",
	itemStatus:             "itemStatus",
	itemAccountName:        "itemAccountName",
	itemAccount:            "itemAccount",
	itemBeginAutomatedXact: "itemBeginAutomatedXact",
	itemBeginPeriodicXact:  "itemBeginPeriodicXact",
	itemBeginXact:          "itemBeginXact",
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

const (
	spaceChars = " \t"
)

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	it := item{t, l.start, l.input[l.start:l.pos]}
	//debug(fmt.Sprintf("Piping item: %v", it))
	l.items <- it
	l.start = l.pos
}

func (l *lexer) current() string {
	return l.input[l.start:l.pos]
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	item := <-l.items
	//debug(fmt.Sprintf("Reading item %s", item))
	l.lastPos = item.pos
	return item
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexJournal; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

// Lex State Functions

// lexJournal scans the Ledger file for top-level Ledger constructs.
func lexJournal(l *lexer) stateFn {
	switch r := l.next(); {
	case isSpace(r):
		l.emitSpaces()
	case r == '~':
		l.emit(itemTilde)
		return lexPeriodicXact
	case r == '=':
		l.emit(itemEqual)
		return lexAutomatedXact
	case unicode.IsDigit(r):
		l.backup()
		return lexPlainXact
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case isEndOfLine(r):
		l.emit(itemEOL)
	case isComment(r):
		l.backup()
		l.emitCommentToEOL()
	case r == eof:
		l.emit(itemEOF)
		return nil
	default:
		return l.errorf("unrecognized character in directive: %#U", r)
	}
	return lexJournal
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch {
			case word == "include":
				l.emit(itemInclude)
				l.emitSpaces()
				if !l.emitStringToEOL() {
					l.errorf("missing filename after 'include'")
					return nil
				}
			case word == "P":
			case r == 'P':
				if !isSpace(l.peek()) {
					return l.errorf("directive 'P' must be followed by a space")
				}
				l.emitSpaces()

				return lexPriceDirective

			case word == "end":
				l.emit(itemEnd)
				l.emitSpaces()
				return lexIdentifier
				// handle "alias", etc..
			case key[word] > itemKeyword:
				l.emit(key[word])
			default:
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}
	return lexJournal
}

func lexPriceDirective(l *lexer) stateFn {
	return l.errorf("price directive not yet supported")
}

func lexPeriodicXact(l *lexer) stateFn {
	// TODO: support those things...
	l.emitSpaces()
	l.emitStringNote()
	return lexPostings
}

func lexAutomatedXact(l *lexer) stateFn {
	// TODO: support those things...
	l.emitSpaces()
	l.emitStringNote()
	return lexPostings
}

func lexPlainXact(l *lexer) stateFn {
	if !l.scanDate() {
		return nil
	}
	l.emit(itemDate)
	l.emitSpaces()

	switch r := l.peek(); {
	case r == eof:
		l.errorf("unexpected end-of-file")
		return nil
	case isEndOfLine(r):
		l.errorf("unexpected end-of-line")
		return nil
	case r == '=':
		l.next()
		l.emit(itemEqual)
		l.emitSpaces()
		if !l.scanDate() {
			return nil
		}
		l.emit(itemDate)
		l.emitSpaces()
	}
	return lexPlainXactDescription
}

func lexPlainXactDescription(l *lexer) stateFn {
	switch r := l.next(); {
	case r == '*':
		l.emit(itemAsterisk)
		l.emitSpaces()
	case r == '!':
		l.emit(itemExclamation)
		l.emitSpaces()
	case r == '(':
		l.emit(itemLeftParen)
		l.scanStringUntil(')')
	case r == ')':
		l.emit(itemRightParen)
		l.emitSpaces()
	case isEndOfLine(r):
		l.errorf("unexpected end-of-line")
		return nil
	case r == eof:
		l.errorf("unexpected end-of-file")
		return nil
	default:
		l.backup()
		l.emitStringNote()
		return lexPostings
	}
	return lexPlainXactDescription
}

func (l *lexer) emitSpaces() bool {
	for isSpace(l.peek()) {
		l.next()
	}
	if l.current() == "" {
		return false
	}
	l.emit(itemSpace)
	return true
}

func (l *lexer) emitStringToEOL() bool {
	if !l.scanStringToEOL() {
		return false
	}
	l.emit(itemString)
	return true
}

func (l *lexer) emitCommentToEOL() bool {
	if !l.scanStringToEOL() {
		return false
	}
	l.emit(itemComment)
	return true
}

func (l *lexer) scanStringToEOL() bool {
Loop:
	for {
		switch r := l.peek(); {
		case isEndOfLine(r) || r == eof:
			break Loop
		default:
			l.next()
		}
	}
	if l.current() == "" {
		return false
	}
	return true
}

func (l *lexer) emitNote() {
	for {
		switch r := l.next(); {
		case isEndOfLine(r) || r == eof:
			l.backup()
			if l.current() != "" {
				l.emit(itemNote)
			}
			return
		}
	}
}

func (l *lexer) emitStringNote() {
Loop:
	for {
		switch r := l.next(); {
		case r == ';':
			l.backup()
			if l.current() != "" {
				l.emit(itemString)
			}
			l.emitNote()
			break Loop
		case isEndOfLine(r):
			l.backup()
			if l.current() != "" {
				l.emit(itemString)
			}
			break Loop
		case r == eof:
			l.backup()
			if l.current() != "" {
				l.emit(itemString)
			}
			break Loop
		default:
			continue
		}
	}
	return
}

// scanAccountName scans until a spacer ("  " or " \t" or "\t " or "\t")
func (l *lexer) scanAccountName() bool {
Loop:
	for {
		switch r := l.peek(); {
		case isEndOfLine(r):
			break Loop
		case r == eof:
			break Loop
		case r == '\t':
			break Loop
		case r == ' ':
			l.next()
			if r2 := l.peek(); r2 == ' ' || r2 == '\t' {
				l.backup()
				break Loop
			}
		default:
			l.next()
		}
	}
	if l.current() == "" {
		l.errorf("expected account name after leading spacesempty account name")
		return false
	}
	l.emit(itemAccountName)

	return true
}

func (l *lexer) scanStringUntil(until rune) bool {
Loop:
	for {
		switch r := l.peek(); {
		case isEndOfLine(r):
			break Loop
		case r == eof:
			break Loop
		case r == until:
			break Loop
		default:
			l.next()
		}
	}
	if l.current() == "" {
		return false
	}
	l.emit(itemString)
	return true
}

func lexPostings(l *lexer) stateFn {
	// Always arrive here with an EOL as first token, or an Account name directly.
	var expectIndent bool

Loop:
	for {
		r := l.next()
		if expectIndent {
			if isSpace(r) {
				l.emitSpaces()
				break Loop
			}
			l.backup()
			return lexJournal
		}

		switch {
		case isEndOfLine(r):
			expectIndent = true
			l.emit(itemEOL)
			continue
		case r == eof:
			l.backup()
			return lexJournal
		case r == '*':
			l.emit(itemAsterisk)
			l.emitSpaces()
		case r == '!':
			l.emit(itemExclamation)
			l.emitSpaces()
		case isComment(r):
			l.backup()
			l.emitNote()
		case unicode.IsLetter(r):
			if !l.scanAccountName() {
				return nil
			}
			l.emitSpaces()
			return lexPostingValues
		}
		expectIndent = false
	}
	return lexPostings
}

func lexPostingValues(l *lexer) stateFn {
	r := l.peek()
	//debug(fmt.Sprintf("lexValues %s %v", string(r), r))
	switch {
	case r == '(':
		l.next()
		if !l.emitValueExpr() {
			return nil
		}
	case r == '=':
		l.next()
		l.emit(itemEqual)
	case r == '-':
		l.next()
		l.emit(itemNeg)
	case unicode.IsDigit(r) || r == '.':
		if !l.emitQuantity() {
			return nil
		}
	case r == '@':
		l.emitPrices()
	case r == '[':
		l.next()
		if !l.scanDate() {
			return nil
		}
		if !l.accept("]") {
			l.errorf("expected matching ']' for lot date, got %#U", l.next())
			return nil
		}
		l.emit(itemLotDate)
	case r == '{':
		// TODO: this should support a full "amount" to be specified,
		// so "- 123 USD", not only a quantity.
		// We should split this function `lexPostingValues` into `lexAmount`
		// which would only handle the amounts, so we could feed it more than once.
		l.next()
		if !l.scanQuantity() {
			return nil
		}
		if !l.accept("}") {
			l.errorf("expected matching '}' for lot price, got %#U, don't specify commodity here", l.next())
			return nil
		}
		l.emit(itemLotPrice)
		// TODO: currently no support for '(' lot_note ')'..
	case isSpace(r):
		l.emitSpaces()
	case isEndOfLine(r):
		return lexPostings
	case r == eof:
		return lexPostings
	case r == ';':
		l.emitNote()
	case isCommodity(r):
		if !l.scanCommodity() {
			return nil
		}
		l.emit(itemCommodity)
	default:
		l.errorf("unexpected character in posting values: %#U", r)
		return nil
	}
	return lexPostingValues
	/*
		HERE SCAN: values_opt note_opt EOL

		values_opt:
		    spacer amount_expr price_opt |
		    [epsilon]
		    ;

		amount_expr: amount | value_expr ;

		amount:
		    neg_opt commodity quantity annotation |
		    quantity commodity annotation ;

		price_opt: price | [epsilon] ;
		price:
		    '@' amount_expr |
		    '@@' amount_expr            [in this case, it's the whole price]
		    ;

		annotation: lot_price_opt lot_date_opt lot_note_opt ;

		lot_date_opt: date | [epsilon] ;
		lot_date: '[' date ']' ;

		lot_price_opt: price | [epsilon] ;
		lot_price: '{' amount '}' ;

		lot_note_opt: note | [epsilon] ;
		lot_note: '(' string ')' ;

	*/
}

var debugCount = 0

func debug(text string) {
	debugCount++
	if debugCount < 100 {
		fmt.Println(text)
	}
}

func (l *lexer) scanCommodity() bool {
	quotesOpen := false
	for {
		switch r := l.next(); {
		case r == '\\':
			r2 := l.next()
			if r2 == eof || isEndOfLine(r2) {
				l.errorf("unexpected end of escape sequence")
				return false
			}
		case unicode.IsDigit(r) || r == ' ' || r == '-' || r == '.':
			if !quotesOpen {
				l.backup()
				return true
			}
		case r == '"':
			if quotesOpen {
				return true
			}
			quotesOpen = true
		case isEndOfLine(r) || r == eof:
			l.backup()
			return true
		}
	}
}

// emitPrices analyzes the '@' and '@@' syntax in posting values.
func (l *lexer) emitPrices() {
	l.next() // consume the @ which brought us here
	if l.peek() == '@' {
		l.next()
		l.emit(itemDoubleAt)
	} else {
		l.emit(itemAt)
	}
	l.emitSpaces()
}

// scanDate scans dates in whatever format.
func (l *lexer) scanDate() bool {
	const dateError = "date format error, expects YYYY-MM-DD with '/', '-' or '.' as separators, received character %#U"
	fields := []int{4, 2, 2}
	for {
		fieldExpected := fields[0]

		switch r := l.next(); {
		case unicode.IsDigit(r):
			fieldExpected--
			fields[0] = fieldExpected
			if fieldExpected < 0 {
				l.errorf(dateError, r)
				return false
			}
		case r == '.' || r == '-' || r == '/':
			if fieldExpected != 0 {
				l.errorf(dateError, r)
				return false
			}
			if len(fields) == 0 {
				l.errorf(dateError, r)
				return false
			}
			fields = fields[1:]
		default:
			if fieldExpected != 0 || len(fields) != 1 {
				l.errorf(dateError, r)
				return false
			}
			l.backup()
			return true
		}
	}
}

// emitValueExpr reads until last unbound ')'
func (l *lexer) emitValueExpr() bool {
	var parenCount int
	for {
		switch r := l.next(); {
		case r == '(':
			parenCount++
		case r == ')':
			parenCount--
			if parenCount < 0 {
				l.emit(itemValueExpr)
				return true
			}
		case isEndOfLine(r) || r == eof:
			l.errorf("unexpected end of amount expression, expected ')'")
			return false
		}
	}
}

// emitQuantity picks up numbers, with decimal points and commas. No
// scientific notation, complex numbers, etc.. Negativity is picked up
// by itemNeg by the caller.
func (l *lexer) emitQuantity() bool {
	if !l.scanQuantity() {
		return false
	}
	l.emit(itemQuantity)
	return true
}

func (l *lexer) scanQuantity() bool {
	for {
		switch r := l.next(); {
		case unicode.IsDigit(r) || r == '.' || r == ',':
			// consume...
		case r == '\t' || r == ' ' || r == '}' || isEndOfLine(r) || r == eof || isCommodity(r):
			l.backup()
			return true
		default:
			l.errorf("invalid character in amount: %#U", r)
			return false
		}
	}
}

func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	return false
}

// isCommodity reports whether r is a valid commodity character, like
// "U" from "USD", '"' like "pine apples" or "$" et al.
func isCommodity(r rune) bool {
	if r == ';' {
		return false
	}
	if unicode.IsGraphic(r) {
		return true
	}
	return false
}

func isComment(r rune) bool {
	return r == ';' || r == '#' || r == '%' || r == '|' || r == '*'
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
