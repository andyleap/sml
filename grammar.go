package sml

import (
	"fmt"
	"io"

	. "github.com/andyleap/stateparser"
)

type RuneSeeker interface {
	io.Seeker
	io.RuneReader
}

type IndentStateReader struct {
	rs     RuneSeeker
	indent int
}

type IndentState struct {
	Indent int
	pos    int64
}

func (isr *IndentStateReader) ReadRune() (rune, int, error) {
	r, size, err := isr.rs.ReadRune()
	if err != nil {
		return r, size, err
	}
	isr.indent++
	if r == '\n' {
		isr.indent = 0
	}
	return r, size, err
}

func (isr *IndentStateReader) State() interface{} {
	pos, _ := isr.rs.Seek(0, io.SeekCurrent)
	return IndentState{
		Indent: isr.indent,
		pos:    pos,
	}
}

func (isr *IndentStateReader) RestoreState(state interface{}) {
	indentState := state.(IndentState)
	isr.indent = indentState.Indent
	isr.rs.Seek(indentState.pos, io.SeekStart)
}

func debug(sr StateReader, n int) {
	state := sr.State()
	r := make([]rune, n)
	var err error
	for i := range r {
		r[i], _, err = sr.ReadRune()
		if err != nil {
			sr.RestoreState(state)
			return
		}
	}
	sr.RestoreState(state)
}

func block(element Grammar) Grammar {
	whitespace := Mult(0, 0, Set(" \t"))
	return func(sr StateReader) (interface{}, error) {
		_, err := whitespace(sr)
		if err != nil {
			return nil, err
		}
		indent := (sr.State().(IndentState)).Indent
		es := []interface{}{}
		for {
			debug(sr, 5)
			e, err := element(sr)
			if err != nil {
				if len(es) == 0 {
					return nil, fmt.Errorf("No items")
				}
				return es, nil
			}
			es = append(es, e)
			state := sr.State()
			_, err = whitespace(sr)
			if err != nil {
				return es, nil
			}
			nextIndent := (sr.State().(IndentState)).Indent
			if nextIndent != indent {
				sr.RestoreState(state)
				return es, nil
			}
		}
	}
}

func nlOrEOF(sr StateReader) (interface{}, error) {
	state := sr.State()
	r, _, err := sr.ReadRune()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		sr.RestoreState(state)
		return nil, err
	}
	if r != '\n' {
		sr.RestoreState(state)
		return nil, fmt.Errorf("Expected \"\\n\", got %q", r)
	}
	return "\n", nil
}

func makeGrammar() Grammar {
	ws := Mult(0, 0, Set(" \t"))

	ident := Node(Mult(1, 0, Set("^ \t\n:")), func(m interface{}) (interface{}, error) {
		return String(m), nil
	})
	literal := Node(And(Tag("value", Mult(1, 0, Set("^\n"))), nlOrEOF), func(m interface{}) (interface{}, error) {
		return String(GetTag(m, "value")), nil
	})

	v1 := Lit("")
	array := &v1
	v2 := Lit("")
	mapBlock := &v2

	*array = Node(And(block(
		And(Lit("-"), ws, Tag("item", Or(Resolve(array), Resolve(mapBlock), literal))),
	)), func(m interface{}) (interface{}, error) {
		ms := GetTags(m, "item")
		items := []interface{}{}
		for _, subm := range ms {
			items = append(items, subm)
		}
		return items, nil
	})

	mapItem := And(
		Tag("key", ident),
		Lit(":"),
		ws,
		Or(And(Lit("\n"), Tag("value", Resolve(array))), And(Lit("\n"), Tag("value", Resolve(mapBlock))), Tag("value", literal)),
	)

	*mapBlock = Node(And(block(Tag("item", mapItem))), func(m interface{}) (interface{}, error) {
		ms := GetTags(m, "item")
		ret := map[string]interface{}{}
		for _, subm := range ms {
			ret[GetTag(subm, "key").(string)] = GetTag(subm, "value")
		}
		return ret, nil
	})

	return Or(Resolve(array), Resolve(mapBlock), literal)
}
