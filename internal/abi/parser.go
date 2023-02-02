package abi

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var ErrSyntax = errors.New("syntax error")

// ParseArgs parses the given Solidity args and returns its arguments.
func ParseArgs(s string) (a Arguments, err error) {
	l := newLexer(s)
	p := newParser(l)

	if err := p.parseArgs(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSyntax, err)
	}
	return (Arguments)(p.args), nil
}

// ParseArgsWithName parses the given Solidity function/event signature and
// returns its name and arguments.
func ParseArgsWithName(s string) (name string, a Arguments, err error) {
	l := newLexer(s)
	p := newParser(l)

	if err := p.parseArgsWithName(); err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrSyntax, err)
	}
	return p.name, (Arguments)(p.args), nil
}

type parser struct {
	lexer *lexer
	items []*item
	i     int

	name string
	args abi.Arguments

	err error
}

func newParser(lexer *lexer) *parser {
	return &parser{
		lexer: lexer,
		items: make([]*item, 0),
		i:     -1,
	}
}

func (p *parser) next() *item {
	if p.i < len(p.items)-1 {
		p.i += 1
		return p.items[p.i]
	}

	next, err := p.lexer.nextItem()
	if err != nil {
		p.err = err
		return nil
	}

	p.i += 1
	p.items = append(p.items, next)
	return next
}

func (p *parser) backup() {
	p.i -= 1
}

func (p *parser) peek() *item {
	next := p.next()
	p.backup()
	return next
}

func (p *parser) parseArgsWithName() error {
	// parse name
	if next := p.next(); next.Typ != itemTypeID {
		return fmt.Errorf(`unexpected %s, expecting name`, next)
	} else {
		p.name = next.Val
	}

	// parse "("
	if next := p.next(); next.Typ != itemTypePunct || next.Val != "(" {
		return fmt.Errorf(`unexpected %s, expecting "("`, next)
	}

	if err := p.parseArgs(); err != nil {
		return err
	}

	// parse ")"
	if next := p.next(); next.Typ != itemTypePunct || next.Val != ")" {
		return fmt.Errorf(`unexpected %s, expecting ")"`, next)
	}

	// parse EOF
	if next := p.next(); next.Typ != itemTypeEOF {
		return fmt.Errorf(`unexpected %s, expecting EOF`, next)
	}
	return nil
}

func (p *parser) parseArgs() error {
	if peek := p.peek(); (peek.Typ == itemTypeEOF && p.name == "") ||
		(peek.Typ == itemTypePunct && peek.Val == ")" && p.name != "") {
		return nil
	}

	for {
		// parse type
		typ, err := p.parseType()
		if err != nil {
			return err
		}
		arg := abi.Argument{Type: *typ}

		// parse optional indexed and name
		peek := p.peek()
		if peek.Typ == itemTypeID {
			if peek.Val == "indexed" {
				arg.Indexed = true
			} else {
				arg.Name = peek.Val
			}
			p.next()

			peek = p.peek()
			if peek.Typ == itemTypeID && arg.Indexed {
				arg.Name = peek.Val
				p.next()
			}
		}

		p.args = append(p.args, arg)

		// parse ",", EOF, or ")"
		if peek := p.peek(); peek.Typ == itemTypeEOF && p.name == "" {
			break
		} else if peek.Typ == itemTypePunct && peek.Val == ")" && p.name != "" {
			break
		} else if peek.Typ == itemTypePunct && peek.Val == "," {
			p.next()
		} else {
			if p.name == "" {
				return fmt.Errorf(`unexpected %s, want "," or EOF`, peek)
			} else {
				return fmt.Errorf(`unexpected %s, want "," or ")"`, peek)
			}
		}
	}
	return nil
}

// parseType parses a non-tupple type of form "type (indexed)? (name)?"
func (p *parser) parseType() (*abi.Type, error) {
	var (
		typ *abi.Type
		ok  bool
		err error
	)
	if peek := p.peek(); peek.Typ == itemTypeID {
		// non-tuple type
		typ, ok = peek.IsType()
		if !ok {
			return nil, fmt.Errorf(`unexpected %s, expecting type`, peek)
		}
		p.next()
	} else if peek.Typ == itemTypePunct && peek.Val == "(" {
		// tuple type
		typ, err = p.parseTupleTypes()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf(`unexpected %s, expecting type`, peek)
	}

	// optional: parse slice or array
	typ, err = p.parseSliceOrArray(typ)
	if err != nil {
		return nil, err
	}

	return typ, nil
}

func (p *parser) parseTupleType() (*abi.Type, string, error) {
	typ, err := p.parseType()
	if err != nil {
		return nil, "", err
	}

	// parse name
	next := p.next()
	if next.Typ != itemTypeID {
		return nil, "", fmt.Errorf(`unexpected %s, want name`, next)
	}

	return typ, next.Val, nil
}

func (p *parser) parseTupleTypes() (*abi.Type, error) {
	if next := p.next(); next.Typ != itemTypePunct || next.Val != "(" {
		return nil, fmt.Errorf(`unexpected %s, expecting "("`, next)
	}

	typ := &abi.Type{T: abi.TupleTy}
	fields := make([]reflect.StructField, 0)
	for {
		// parse type
		elemTyp, name, err := p.parseTupleType()
		if err != nil {
			return nil, err
		}
		typ.TupleElems = append(typ.TupleElems, elemTyp)
		typ.TupleRawNames = append(typ.TupleRawNames, name)
		fields = append(fields, reflect.StructField{
			Name: abi.ToCamelCase(name),
			Type: elemTyp.GetType(),
			Tag:  reflect.StructTag(`abi:"` + name + `"`),
		})

		next := p.next()
		if next.Typ == itemTypePunct {
			if next.Val == ")" {
				break
			} else if next.Val == "," {
				continue
			}
		}
		return nil, fmt.Errorf(`unexpected %s, expecting "," or ")"`, next)
	}
	typ.TupleType = reflect.StructOf(fields)
	return typ, nil
}

func (p *parser) parseSliceOrArray(typ *abi.Type) (*abi.Type, error) {
	parent := *typ
	for peek := p.peek(); peek.Typ == itemTypePunct && peek.Val == "["; peek = p.peek() {
		// parse "["
		p.next()

		// nest type
		parentCopy := parent
		parent = abi.Type{
			Elem: &parentCopy,
		}

		// parse optional number
		next := p.next()
		if next.Typ == itemTypeNum {
			parent.Size, _ = strconv.Atoi(next.Val)
			parent.T = abi.ArrayTy
			next = p.next()
		} else {
			parent.T = abi.SliceTy
		}

		// parse "]"
		if next.Typ != itemTypePunct || next.Val != "]" {
			return nil, fmt.Errorf(`unexpected %s, expecting "]"`, next)
		}
	}
	return &parent, nil
}
