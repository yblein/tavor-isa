// Lexer adapted from the Go standard library:
// http://golang.org/src/text/template/parse/lex.go

package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pos int

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
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota

	itemText
	itemInteger
	itemLabel
	itemKey

	itemNewLine
	itemEOF
)

const eof = -1

const comment = '#'

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input   string    // the string being scanned
	state   stateFn   // the next lexing function to enter
	pos     Pos       // current position in the input
	start   Pos       // start position of this item
	width   Pos       // width of last rune read from input
	lastPos Pos       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
}

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

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) int {
	r := 0
	for strings.IndexRune(valid, l.next()) >= 0 {
		r++
	}
	l.backup()
	return r
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
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	l.state = lexText
	for l.state != nil {
		l.state = l.state(l)
	}
	close(l.items)
}

// lexText scans until an special character
func lexText(l *lexer) stateFn {
loop:
	for {
		switch r := l.next(); {
		case r == '$':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.next()
			return lexSpecial
		case r == '@':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.next()
			return lexKey
		case r == comment:
			return lexComment
		case isEndOfLine(r):
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.next()
			l.emit(itemNewLine)
			return lexText
		case r == eof:
			break loop
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

// lexKey scans the content of a key where the @ mark is already scanned.
func lexKey(l *lexer) stateFn {
	for unicode.IsLetter(l.next()) {
	}
	l.backup()
	if l.pos <= l.start+1 {
		return l.errorf("expected a key string after @ character")
	}
	l.emit(itemKey)
	return lexText
}

// lexKey scans the content of a special token where the $ mark is already scanned.
func lexSpecial(l *lexer) stateFn {
	switch l.next() {
	case 'i', 'u':
		if l.acceptRun("0123456789") <= 0 {
			return l.errorf("expected integer size after a $i or $u sequence")
		}
		l.emit(itemInteger)
	case 'l':
		l.emit(itemLabel)
	default:
		return l.errorf("expected 'i', 'u' or 'l' after $ character")
	}
	return lexText
}

// lexComment scans a comment. The left comment marker is already scanned.
func lexComment(l *lexer) stateFn {
	i := strings.Index(l.input[l.pos:], "\n")

	// stop here if this is last line
	if i < 0 {
		l.start = Pos(len(l.input) - 1)
		l.pos = Pos(len(l.input) - 1)
		l.emit(itemEOF)
		return nil
	}

	l.pos += Pos(i + 1)
	l.ignore()
	return lexText
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
