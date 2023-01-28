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

// IsType returns whether the item is a type or not.
func (i *item) IsType() (string, bool) {
	if i.Typ != itemTypeID {
		return "", false
	}
	typ, ok := types[i.Val]
	return typ, ok
}

const (
	// lexer item types
	itemTypeID    itemType = iota // identifier: [a-zA-Z_][0-9a-zA-Z_]*
	itemTypePunct                 // punctuation: [\(\)\[\],]
	itemTypeNum                   // number: [1-9][0-9]*
	itemTypeEOF                   // end of file

	eof rune = -1

	num0  = "123456789"
	num   = "0" + num0
	id0   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	id    = id0 + num
	space = " \t\r\n"
)

// lexer holds the state of the scanner.
type lexer struct {
	input string // the string being scanned
	start int    // start position of this item
	pos   int    // current position in the input
	width int    // width of last rune read from input
}

func newLexer(input string) *lexer {
	return &lexer{input: input}
}

func (l *lexer) nextItem() (*item, error) {
Start:
	switch l.peek() {
	case ' ', '\t', '\r', '\n':
		l.acceptRun(space)
		l.ignore()
		goto Start
	case '(', ')', '[', ']', ',':
		l.next()
		return &item{itemTypePunct, l.token()}, nil
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		l.accept(num0)
		l.acceptRun(num)
		return &item{itemTypeNum, l.token()}, nil
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '_':
		l.accept(id0)
		l.acceptRun(id)
		return &item{itemTypeID, l.token()}, nil
	case eof:
		return &item{itemTypeEOF, ""}, nil
	default:
		return nil, fmt.Errorf("unexpected character: %c", l.next())
	}
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
	token := l.input[l.start:l.pos]
	l.ignore()
	return token
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
