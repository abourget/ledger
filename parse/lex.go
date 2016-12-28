package parse

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
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemString
	itemNote // sort of Comment for postings, etc..
	itemJournal
	itemJournalItem
	itemDate
	itemDirective
	itemXact
	itemSpace
	itemEOL
	itemText
	itemAsterisk
	itemSemicolon
	itemCommodity
	itemAmount
	itemIdentifier
	itemNeg
	itemInteger
	itemTilde
	itemPeriodExpr
	itemDot // to form numbers, with itemInteger + optionally: itemDot + itemInteger
	itemStatus
	itemAccountName // only a name like "Expenses:Misc"
	itemAccount     // can be "(Expenses:Misc)" or "[Expenses:Misc]"
	itemBeginAutomaticXact
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
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
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
		for isSpace(l.peek()) {
			l.next()
		}
		l.emit(itemSpace)
	case r == '~':
		l.emit(itemBeginPeriodicXact)
		return lexPeriodicXact
	case r == '=':
		l.emit(itemBeginAutomaticXact)
		return lexAutomaticXact
	case unicode.IsDigit(r):
		return lexPlainXact
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case isEndOfLine(r):
		l.emit(itemEOL)
	case r == eof:
		l.emit(itemEOF)
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
				l.scanSpaces()
				if !l.scanStringToEOL() {
					l.errorf("missing filename after 'include'")
					return nil
				}
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

func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	return false
}

func lexPeriodicXact(l *lexer) stateFn {
	l.scanSpaces()
Loop:
	for {
		switch r := l.next(); {
		case r == ';':
			l.backup()
			if l.current() == "" {
				l.errorf("missing period expression")
				return nil
			}
			l.emit(itemPeriodExpr)
			l.next() // skip that ';' again..
			l.scanNote()
			break Loop
		case isEndOfLine(r):
			l.emit(itemPeriodExpr)
			break
		case r == eof:
			l.errorf("premature end-of-file, expected postings for periodic transaction")
			return nil
		default:
			continue
		}
	}
	return lexPostings
}

func (l *lexer) scanNote() {
	for {
		switch r := l.next(); {
		case isEndOfLine(r) || r == eof:
			l.emit(itemNote)
			return
		default:
			continue
		}
	}

}

func lexAutomaticXact(l *lexer) stateFn {
	l.errorf("not yet implemented")
	return nil
}

func lexPlainXact(l *lexer) stateFn {
	for {
		// Refine the `date` parsing at the Lex-level..
		switch r := l.peek(); {
		case unicode.IsDigit(r):
			l.next()
		case r == '.' || r == '-' || r == '/':
			l.next()
		case isSpace(r):
			l.emit(itemDate)
			l.scanSpaces()
			return lexPostings
		case r == eof:
			l.errorf("unexpected end-of-file, expected transaction Payee or Description")
			return nil
		case isEndOfLine(r):
			l.errorf("unexpected end-of-line, expected transaction Payee or Description")
			return nil
		default:
			l.errorf("invalid character in transaction date specification: %q", r)
			return nil
		}
	}
}

func (l *lexer) scanSpaces() bool {
	for isSpace(l.peek()) {
		l.next()
	}
	if l.current() == "" {
		return false
	}
	l.emit(itemSpace)
	return true
}

func (l *lexer) scanStringToEOL() bool {
Loop:
	for {
		switch r := l.peek(); {
		case isEndOfLine(r):
			break Loop
		case r == eof:
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
	l.errorf("not yet implemented")
	return nil
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
