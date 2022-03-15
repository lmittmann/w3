package abi

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type item struct {
	Typ itemType
	Val string
}

type itemType int

const (
	itemError      itemType = iota
	itemID                  // function or argument name
	itemTyp                 // argument type
	itemLeftParen           // '('
	itemRightParen          // ')'
	itemDelim               // ','
	itemEOF
)

type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input  string      // the string being scanned
	start  int         // start position of this item
	pos    int         // current position in the input
	width  int         // width of last rune read from input
	itemCh chan<- item // channel of scanned items
}

func newLexer(input string, itemCh chan<- item) *lexer {
	return &lexer{
		input:  input,
		itemCh: itemCh,
	}
}

func (l *lexer) run() {
	for state := lexFuncOrType; state != nil; {
		state = state(l)
	}
	close(l.itemCh)
}

func (l *lexer) next() (next rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	next, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) backupAll() {
	l.pos = l.start
	l.width = 0
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() (next rune) {
	next = l.next()
	l.backup()
	return
}

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

func (l *lexer) token() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) emit(typ itemType) {
	l.emitVal(typ, l.token())
}

func (l *lexer) emitVal(typ itemType, val string) {
	l.itemCh <- item{typ, val}
	l.start = l.pos
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state.
func (l *lexer) errorf(format string, args ...any) stateFn {
	l.itemCh <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

const (
	eof rune = -1

	number  = "0123456789"
	idStart = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ$_"
	idPart  = idStart + number
	space   = " \t\r\n"
)

func lexFuncOrType(l *lexer) stateFn {
	switch l.peek() {
	case '(':
		return lexLeftParent
	case eof:
		l.emit(itemEOF)
		return nil
	}

	l.accept(idStart)
	l.acceptRun(idPart)

	switch l.peek() {
	case '(':
		l.emit(itemID)
		return lexLeftParent
	default:
		l.backupAll()
		return lexType
	}
}

func lexLeftParent(l *lexer) stateFn {
	if next := l.next(); next == '(' {
		l.emit(itemLeftParen)
	} else {
		return l.errorf("unexpected token %q, want '('", next)
	}

	switch l.peek() {
	case '(':
		return lexLeftParent
	case ')':
		return lexRightParent
	case eof:
		return l.errorf("unexpected EOF after '('")
	default:
		return lexType
	}
}

func lexRightParent(l *lexer) stateFn {
	if next := l.next(); next == ')' {
		l.emit(itemRightParen)
	} else {
		return l.errorf("unexpected token %q, want ')'", next)
	}

	switch peek := l.peek(); peek {
	case ')':
		return lexRightParent
	case ' ', '\t', '\r', '\n':
		return lexArgName
	case ',':
		return lexDelim
	case eof:
		l.emit(itemEOF)
		return nil
	default:
		return l.errorf("unexpected token %q, want ')', ',' or EOF", peek)
	}
}

func lexType(l *lexer) stateFn {
	// ignore leading spaces
	l.acceptRun(space)
	l.ignore()

	// accept type
	l.accept(idStart)
	l.acceptRun(idPart)
	normType, ok := types[l.token()]
	if !ok {
		return l.errorf("unknown type %q", l.token())
	}
	l.ignore()

	// optionally accept array
	for l.peek() == '[' {
		l.accept("[")
		l.acceptRun(number)
		switch peek := l.peek(); peek {
		case ']':
			l.accept("]")
		case eof:
			return l.errorf("unexpected EOF, want ']'")
		default:
			return l.errorf("unexpected token %q, want ']'", peek)
		}
		normType += l.token()
		l.ignore()
	}
	l.emitVal(itemTyp, normType)

	switch peek := l.peek(); peek {
	case ',':
		return lexDelim
	case ' ', '\t', '\r', '\n':
		return lexArgName
	case ')':
		return lexRightParent
	case eof:
		l.emit(itemEOF)
		return nil
	default:
		return lexType
	}
}

func lexArgName(l *lexer) stateFn {
	// ignore leading spaces
	l.acceptRun(space)
	l.ignore()

	l.accept(idStart)
	l.acceptRun(idPart)
	l.emit(itemID)

	switch peek := l.peek(); peek {
	case ')':
		return lexRightParent
	case ',':
		return lexDelim
	case eof:
		l.emit(itemEOF)
		return nil
	default:
		return l.errorf("unexpected token %q, want ')', ',' or EOF", peek)
	}
}

func lexDelim(l *lexer) stateFn {
	if next := l.next(); next == ',' {
		l.emit(itemDelim)
	} else {
		return l.errorf("unexpected token %q, want ','", next)
	}

	l.acceptRun(space)
	l.ignore()

	switch peek := l.peek(); peek {
	case '(':
		return lexLeftParent
	case eof:
		return l.errorf("unexpected EOF after ','")
	default:
		return lexType
	}
}

var types = map[string]string{
	"address": "address",
	"bool":    "bool",
	"string":  "string",
	"bytes":   "bytes",
	"bytes1":  "bytes1",
	"bytes2":  "bytes2",
	"bytes3":  "bytes3",
	"bytes4":  "bytes4",
	"bytes5":  "bytes5",
	"bytes6":  "bytes6",
	"bytes7":  "bytes7",
	"bytes8":  "bytes8",
	"bytes9":  "bytes9",
	"bytes10": "bytes10",
	"bytes11": "bytes11",
	"bytes12": "bytes12",
	"bytes13": "bytes13",
	"bytes14": "bytes14",
	"bytes15": "bytes15",
	"bytes16": "bytes16",
	"bytes17": "bytes17",
	"bytes18": "bytes18",
	"bytes19": "bytes19",
	"bytes20": "bytes20",
	"bytes21": "bytes21",
	"bytes22": "bytes22",
	"bytes23": "bytes23",
	"bytes24": "bytes24",
	"bytes25": "bytes25",
	"bytes26": "bytes26",
	"bytes27": "bytes27",
	"bytes28": "bytes28",
	"bytes29": "bytes29",
	"bytes30": "bytes30",
	"bytes31": "bytes31",
	"bytes32": "bytes32",
	"int":     "int256",
	"int8":    "int8",
	"int16":   "int16",
	"int24":   "int24",
	"int32":   "int32",
	"int40":   "int40",
	"int48":   "int48",
	"int56":   "int56",
	"int64":   "int64",
	"int72":   "int72",
	"int80":   "int80",
	"int88":   "int88",
	"int96":   "int96",
	"int104":  "int104",
	"int112":  "int112",
	"int120":  "int120",
	"int128":  "int128",
	"int136":  "int136",
	"int144":  "int144",
	"int152":  "int152",
	"int160":  "int160",
	"int168":  "int168",
	"int176":  "int176",
	"int184":  "int184",
	"int192":  "int192",
	"int200":  "int200",
	"int208":  "int208",
	"int216":  "int216",
	"int224":  "int224",
	"int232":  "int232",
	"int240":  "int240",
	"int248":  "int248",
	"int256":  "int256",
	"uint":    "uint256",
	"uint8":   "uint8",
	"uint16":  "uint16",
	"uint24":  "uint24",
	"uint32":  "uint32",
	"uint40":  "uint40",
	"uint48":  "uint48",
	"uint56":  "uint56",
	"uint64":  "uint64",
	"uint72":  "uint72",
	"uint80":  "uint80",
	"uint88":  "uint88",
	"uint96":  "uint96",
	"uint104": "uint104",
	"uint112": "uint112",
	"uint120": "uint120",
	"uint128": "uint128",
	"uint136": "uint136",
	"uint144": "uint144",
	"uint152": "uint152",
	"uint160": "uint160",
	"uint168": "uint168",
	"uint176": "uint176",
	"uint184": "uint184",
	"uint192": "uint192",
	"uint200": "uint200",
	"uint208": "uint208",
	"uint216": "uint216",
	"uint224": "uint224",
	"uint232": "uint232",
	"uint240": "uint240",
	"uint248": "uint248",
	"uint256": "uint256",
}
