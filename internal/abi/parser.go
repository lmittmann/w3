package abi

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func parse(s string) (name string, args abi.Arguments, err error) {
	itemCh := make(chan item, 1)
	l := newLexer(s, itemCh)
	go l.run()

	p := newParser(itemCh)
	err = p.run()
	if err != nil {
		return
	}

	args = p.args
	name = p.fnName
	return
}

type parser struct {
	itemCh <-chan item // channel of scanned items
	items  []item
	i      int

	args   abi.Arguments
	fnName string
	depth  int
	err    string
}

func newParser(itemCh <-chan item) *parser {
	return &parser{
		itemCh: itemCh,
		items:  make([]item, 0, 1),
		i:      -1,
	}
}

func (p *parser) run() (err error) {
	switch peek := p.peek(); peek.Typ {
	case itemID:
		err = p.parseFunc()
	case itemTyp:
		err = p.parseArgs(nil)
	case itemError:
		err = fmt.Errorf("lex error: %v", peek.Val)
	case itemEOF:
	default:
		p.err = fmt.Sprintf(`unexpected item %q`, peek)
	}
	return
}

func (p *parser) next() (next item) {
	if p.i < len(p.items)-1 {
		p.i += 1
		next = p.items[p.i]
	} else {
		next = <-p.itemCh
		p.i += 1
		p.items = append(p.items, next)
	}
	return
}

func (p *parser) backup() {
	p.i -= 1
}

func (p *parser) peek() (next item) {
	next = p.next()
	p.backup()
	return
}

func (p *parser) parseFunc() error {
	next := p.next()
	p.fnName = next.Val

	if next := p.next(); next.Typ != itemLeftParen {
		return fmt.Errorf(`unexpected item %q, want "("`, next)
	}

	if err := p.parseArgs(nil); err != nil {
		return err
	}

	if next := p.next(); next.Typ != itemRightParen {
		return fmt.Errorf(`unexpected item %q, want ")"`, next)
	}
	if next := p.next(); next.Typ != itemEOF {
		return fmt.Errorf(`unexpected item %q, want EOF`, next)
	}
	return nil

}

func (p *parser) parseArgs(components *[]abi.ArgumentMarshaling) error {
	p.depth += 1
	defer func() { p.depth-- }()

	for {
		var (
			typ   string
			name  string
			comps []abi.ArgumentMarshaling
		)

		switch peek := p.peek(); peek.Typ {
		case itemTyp:
			typ = p.next().Val
		case itemLeftParen:
			p.next()
			typ = "tuple"
			if err := p.parseArgs(&comps); err != nil {
				return err
			}
			if next := p.next(); next.Typ != itemRightParen {
				return fmt.Errorf(`unexpected item %q, want ")"`, next)
			}
		case itemRightParen:
			return nil
		case itemError:
			return errors.New(p.next().Val)
		default:
			return fmt.Errorf(`unexpected item %q, want type or "("`, p.next().Val)
		}

		if peek := p.peek(); peek.Typ == itemID {
			name = p.next().Val
		}

		if p.depth > 1 {
			*components = append(*components, abi.ArgumentMarshaling{
				Name:       name,
				Type:       typ,
				Components: comps,
			})
		} else {
			ty, err := abi.NewType(typ, "", comps)
			if err != nil {
				return err
			}
			p.args = append(p.args, abi.Argument{
				Name: name,
				Type: ty,
			})
		}

		switch peek := p.peek(); peek.Typ {
		case itemDelim:
			p.next()
		case itemRightParen:
			return nil
		case itemEOF:
			return nil
		}
	}
}
