package parse

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/zimmski/tavor"
	"github.com/zimmski/tavor/token"
	"github.com/zimmski/tavor/token/lists"
	"github.com/zimmski/tavor/token/primitives"
)

// Config represents the configuration of an ISA
type Config struct {
	Instructions []string
	Variables    map[string][]string
}

// Parse parses the given configuration file and returns a Tavor token out of it
func Parse(file string) (token.Token, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var conf Config
	if _, err := toml.Decode(string(buf), &conf); err != nil {
		return nil, err
	}

	// build variables out from the configuration file
	variables := make(map[string]token.Token)
	for k, a := range conf.Variables {
		var l []token.Token
		//for _, s := range a {
		//	l = append(l, primitives.NewConstantString(s))
		//}

		// use only the boundary values (first, mid, last) of the variable for now
		if len(a) >= 1 {
			l = append(l, primitives.NewConstantString(a[0]))
			if len(a) >= 3 {
				l = append(l, primitives.NewConstantString(a[len(a)/2]))
			}
			if len(a) >= 2 {
				l = append(l, primitives.NewConstantString(a[len(a)-1]))
			}
		}
		variables[k] = lists.NewOne(l...)
	}

	dir := filepath.Dir(file)
	var l []token.Token

	// parse all instruction files
	for _, instructions := range conf.Instructions {
		file := filepath.Join(dir, instructions)
		t, err := parseInstructions(file, variables)
		if err != nil {
			return nil, err
		}
		l = append(l, t)
	}

	one := lists.NewOne(l...)
	all := lists.NewAll(one, primitives.NewConstantString("\n"))
	return lists.NewRepeat(all, 1, int64(tavor.MaxRepeat)), nil
}

func parseInstructions(file string, variables map[string]token.Token) (token.Token, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	l := lex(string(buf))

	var instructions []token.Token
	var currInstr []token.Token

	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemNewLine:
			instructions = append(instructions, lists.NewAll(currInstr...))
			currInstr = nil
		case itemText:
			currInstr = append(currInstr, primitives.NewConstantString(i.val))
		case itemInteger:
			signed := i.val[1] == 'i'
			nbBits, _ := strconv.Atoi(i.val[2:])
			var from, to int
			if signed {
				from = -(1 << (uint(nbBits) - 1))
				to = (1 << (uint(nbBits) - 1)) - 1
			} else {
				from = 0
				to = (1 << uint(nbBits)) - 1
			}
			currInstr = append(currInstr, primitives.NewRangeInt(from, to))
		case itemLabel:
			currInstr = append(currInstr, primitives.NewConstantString("$l"))
		case itemKey:
			key := i.val[1:]
			if variable, ok := variables[key]; ok {
				currInstr = append(currInstr, variable.Clone())
			} else {
				err := fmt.Errorf("error: %s:%d: variable %s not found", file, l.lineNumber(), key)
				return nil, err
			}
		case itemError:
			err := fmt.Errorf("error: %s:%d: %s", file, l.lineNumber(), i.val)
			return nil, err
		}
	}

	if len(currInstr) > 0 {
		instructions = append(instructions, lists.NewAll(currInstr...))
	}

	return lists.NewOne(instructions...), nil
}
