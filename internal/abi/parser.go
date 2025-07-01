package abi

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var ErrSyntax = errors.New("syntax error")

// Parse parses the given Solidity args and returns its arguments.
func Parse(s string, tuples ...any) (a Arguments, err error) {
	l := newLexer(s)
	p := newParser(l, tuples)

	if err := p.parseArgs(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSyntax, err)
	}
	return (Arguments)(p.args), nil
}

// ParseWithName parses the given Solidity function/event signature and returns
// its name and arguments.
func ParseWithName(s string, tuples ...any) (name string, a Arguments, err error) {
	l := newLexer(s)
	p := newParser(l, tuples)

	if err := p.parseArgsWithName(); err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrSyntax, err)
	}
	return p.name, (Arguments)(p.args), nil
}

type parser struct {
	lexer *lexer
	items []*item
	i     int

	tuples   []any
	tupleMap map[string]abi.Argument

	name string
	args abi.Arguments

	err error
}

func newParser(lexer *lexer, tuples []any) *parser {
	return &parser{
		lexer:  lexer,
		items:  make([]*item, 0),
		i:      -1,
		tuples: tuples,
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
	if next := p.next(); p.err != nil {
		return p.err
	} else if next.Typ != itemTypeID {
		return fmt.Errorf(`unexpected %s, expecting name`, next)
	} else {
		p.name = next.Val
	}

	// parse "("
	if next := p.next(); p.err != nil {
		return p.err
	} else if next.Typ != itemTypePunct || next.Val != "(" {
		return fmt.Errorf(`unexpected %s, expecting "("`, next)
	}

	if err := p.parseArgs(); err != nil {
		return err
	}

	// parse ")"
	if next := p.next(); p.err != nil {
		return p.err
	} else if next.Typ != itemTypePunct || next.Val != ")" {
		return fmt.Errorf(`unexpected %s, expecting ")"`, next)
	}

	// parse EOF
	if next := p.next(); p.err != nil {
		return p.err
	} else if next.Typ != itemTypeEOF {
		return fmt.Errorf(`unexpected %s, expecting EOF`, next)
	}
	return nil
}

func (p *parser) parseArgs() error {
	// parse tuples
	tupleMap, err := buildTuples(p.tuples...)
	if err != nil {
		return err
	}
	p.tupleMap = tupleMap

	if peek := p.peek(); p.err != nil {
		return p.err
	} else if (peek.Typ == itemTypeEOF && p.name == "") ||
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
		if p.err != nil {
			return p.err
		} else if peek.Typ == itemTypeID {
			if peek.Val == "indexed" {
				arg.Indexed = true
			} else {
				arg.Name = peek.Val
			}
			p.next()

			peek = p.peek()
			if p.err != nil {
				return p.err
			} else if peek.Typ == itemTypeID && arg.Indexed {
				arg.Name = peek.Val
				p.next()
			}
		}

		p.args = append(p.args, arg)

		// parse ",", EOF, or ")"
		if peek := p.peek(); p.err != nil {
			return p.err
		} else if peek.Typ == itemTypeEOF && p.name == "" {
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

// parseType parses a non-tuple type of form "type (indexed)? (name)?"
func (p *parser) parseType() (*abi.Type, error) {
	var (
		typ *abi.Type
		ok  bool
		err error
	)
	if peek := p.peek(); p.err != nil {
		return nil, p.err
	} else if peek.Typ == itemTypeID {
		// check built-in types first
		typ, ok = peek.IsType()
		if !ok {
			// check named tuples
			if tupleArg, exists := p.tupleMap[peek.Val]; exists {
				typ = &tupleArg.Type
				ok = true
			}
		}
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

func (p *parser) parseTupleType(i int) (*abi.Type, string, error) {
	typ, err := p.parseType()
	if err != nil {
		return nil, "", err
	}

	// parse name
	next := p.next()
	if p.err != nil {
		return nil, "", p.err
	}
	if next.Typ != itemTypeID {
		// no name given; put the token back and make up a fake name
		p.backup()
		return typ, fmt.Sprintf("arg%d", i), nil
	}

	return typ, next.Val, nil
}

func (p *parser) parseTupleTypes() (*abi.Type, error) {
	if next := p.next(); p.err != nil {
		return nil, p.err
	} else if next.Typ != itemTypePunct || next.Val != "(" {
		return nil, fmt.Errorf(`unexpected %s, expecting "("`, next)
	}

	typ := &abi.Type{T: abi.TupleTy}
	fields := make([]reflect.StructField, 0)
	for i := 0; ; i++ {
		// parse type
		elemTyp, name, err := p.parseTupleType(i)
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
		if p.err != nil {
			return nil, p.err
		}
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
	for peek := p.peek(); p.err == nil && peek.Typ == itemTypePunct && peek.Val == "["; peek = p.peek() {
		// parse "["
		p.next()

		// nest type
		parentCopy := parent
		parent = abi.Type{
			Elem: &parentCopy,
		}

		// parse optional number
		next := p.next()
		if p.err != nil {
			return nil, p.err
		}
		if next.Typ == itemTypeNum {
			parent.Size, _ = strconv.Atoi(next.Val)
			parent.T = abi.ArrayTy
			next = p.next()
			if p.err != nil {
				return nil, p.err
			}
		} else {
			parent.T = abi.SliceTy
		}

		// parse "]"
		if next.Typ != itemTypePunct || next.Val != "]" {
			return nil, fmt.Errorf(`unexpected %s, expecting "]"`, next)
		}
	}
	if p.err != nil {
		return nil, p.err
	}
	return &parent, nil
}
