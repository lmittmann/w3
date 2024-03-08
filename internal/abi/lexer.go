package abi

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type item struct {
	Typ itemType
	Val string
}

type itemType int

func (i item) String() string {
	if i.Typ == itemTypeEOF {
		return "EOF"
	}
	return strconv.Quote(i.Val)
}

// IsType returns whether the item is a type or not.
func (i *item) IsType() (*abi.Type, bool) {
	if i.Typ != itemTypeID {
		return nil, false
	}
	typ, ok := types[i.Val]
	return &typ, ok
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
	case eof, ';':
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

var types = map[string]abi.Type{
	"address": {T: abi.AddressTy, Size: 20},
	"hash":    {T: abi.HashTy, Size: 32},
	"bool":    {T: abi.BoolTy},
	"bytes":   {T: abi.BytesTy},
	"string":  {T: abi.StringTy},
	"bytes1":  {T: abi.FixedBytesTy, Size: 1},
	"bytes2":  {T: abi.FixedBytesTy, Size: 2},
	"bytes3":  {T: abi.FixedBytesTy, Size: 3},
	"bytes4":  {T: abi.FixedBytesTy, Size: 4},
	"bytes5":  {T: abi.FixedBytesTy, Size: 5},
	"bytes6":  {T: abi.FixedBytesTy, Size: 6},
	"bytes7":  {T: abi.FixedBytesTy, Size: 7},
	"bytes8":  {T: abi.FixedBytesTy, Size: 8},
	"bytes9":  {T: abi.FixedBytesTy, Size: 9},
	"bytes10": {T: abi.FixedBytesTy, Size: 10},
	"bytes11": {T: abi.FixedBytesTy, Size: 11},
	"bytes12": {T: abi.FixedBytesTy, Size: 12},
	"bytes13": {T: abi.FixedBytesTy, Size: 13},
	"bytes14": {T: abi.FixedBytesTy, Size: 14},
	"bytes15": {T: abi.FixedBytesTy, Size: 15},
	"bytes16": {T: abi.FixedBytesTy, Size: 16},
	"bytes17": {T: abi.FixedBytesTy, Size: 17},
	"bytes18": {T: abi.FixedBytesTy, Size: 18},
	"bytes19": {T: abi.FixedBytesTy, Size: 19},
	"bytes20": {T: abi.FixedBytesTy, Size: 20},
	"bytes21": {T: abi.FixedBytesTy, Size: 21},
	"bytes22": {T: abi.FixedBytesTy, Size: 22},
	"bytes23": {T: abi.FixedBytesTy, Size: 23},
	"bytes24": {T: abi.FixedBytesTy, Size: 24},
	"bytes25": {T: abi.FixedBytesTy, Size: 25},
	"bytes26": {T: abi.FixedBytesTy, Size: 26},
	"bytes27": {T: abi.FixedBytesTy, Size: 27},
	"bytes28": {T: abi.FixedBytesTy, Size: 28},
	"bytes29": {T: abi.FixedBytesTy, Size: 29},
	"bytes30": {T: abi.FixedBytesTy, Size: 30},
	"bytes31": {T: abi.FixedBytesTy, Size: 31},
	"bytes32": {T: abi.FixedBytesTy, Size: 32},
	"uint8":   {T: abi.UintTy, Size: 8},
	"uint16":  {T: abi.UintTy, Size: 16},
	"uint24":  {T: abi.UintTy, Size: 24},
	"uint32":  {T: abi.UintTy, Size: 32},
	"uint40":  {T: abi.UintTy, Size: 40},
	"uint48":  {T: abi.UintTy, Size: 48},
	"uint56":  {T: abi.UintTy, Size: 56},
	"uint64":  {T: abi.UintTy, Size: 64},
	"uint72":  {T: abi.UintTy, Size: 72},
	"uint80":  {T: abi.UintTy, Size: 80},
	"uint88":  {T: abi.UintTy, Size: 88},
	"uint96":  {T: abi.UintTy, Size: 96},
	"uint104": {T: abi.UintTy, Size: 104},
	"uint112": {T: abi.UintTy, Size: 112},
	"uint120": {T: abi.UintTy, Size: 120},
	"uint128": {T: abi.UintTy, Size: 128},
	"uint136": {T: abi.UintTy, Size: 136},
	"uint144": {T: abi.UintTy, Size: 144},
	"uint152": {T: abi.UintTy, Size: 152},
	"uint160": {T: abi.UintTy, Size: 160},
	"uint168": {T: abi.UintTy, Size: 168},
	"uint176": {T: abi.UintTy, Size: 176},
	"uint184": {T: abi.UintTy, Size: 184},
	"uint192": {T: abi.UintTy, Size: 192},
	"uint200": {T: abi.UintTy, Size: 200},
	"uint208": {T: abi.UintTy, Size: 208},
	"uint216": {T: abi.UintTy, Size: 216},
	"uint224": {T: abi.UintTy, Size: 224},
	"uint232": {T: abi.UintTy, Size: 232},
	"uint240": {T: abi.UintTy, Size: 240},
	"uint248": {T: abi.UintTy, Size: 248},
	"uint256": {T: abi.UintTy, Size: 256},
	"uint":    {T: abi.UintTy, Size: 256},
	"int8":    {T: abi.IntTy, Size: 8},
	"int16":   {T: abi.IntTy, Size: 16},
	"int24":   {T: abi.IntTy, Size: 24},
	"int32":   {T: abi.IntTy, Size: 32},
	"int40":   {T: abi.IntTy, Size: 40},
	"int48":   {T: abi.IntTy, Size: 48},
	"int56":   {T: abi.IntTy, Size: 56},
	"int64":   {T: abi.IntTy, Size: 64},
	"int72":   {T: abi.IntTy, Size: 72},
	"int80":   {T: abi.IntTy, Size: 80},
	"int88":   {T: abi.IntTy, Size: 88},
	"int96":   {T: abi.IntTy, Size: 96},
	"int104":  {T: abi.IntTy, Size: 104},
	"int112":  {T: abi.IntTy, Size: 112},
	"int120":  {T: abi.IntTy, Size: 120},
	"int128":  {T: abi.IntTy, Size: 128},
	"int136":  {T: abi.IntTy, Size: 136},
	"int144":  {T: abi.IntTy, Size: 144},
	"int152":  {T: abi.IntTy, Size: 152},
	"int160":  {T: abi.IntTy, Size: 160},
	"int168":  {T: abi.IntTy, Size: 168},
	"int176":  {T: abi.IntTy, Size: 176},
	"int184":  {T: abi.IntTy, Size: 184},
	"int192":  {T: abi.IntTy, Size: 192},
	"int200":  {T: abi.IntTy, Size: 200},
	"int208":  {T: abi.IntTy, Size: 208},
	"int216":  {T: abi.IntTy, Size: 216},
	"int224":  {T: abi.IntTy, Size: 224},
	"int232":  {T: abi.IntTy, Size: 232},
	"int240":  {T: abi.IntTy, Size: 240},
	"int248":  {T: abi.IntTy, Size: 248},
	"int256":  {T: abi.IntTy, Size: 256},
	"int":     {T: abi.IntTy, Size: 256},
}
