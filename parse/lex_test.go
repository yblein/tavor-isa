package parse

import (
	"testing"

	. "github.com/zimmski/tavor/test/assert"
)

func TestLex(t *testing.T) {
	{
		l := lex("addi @r, @r, $i12\n lb @r, $i12(@r)")
		expected := []item{
			item{typ: itemText, pos: 0, val: "addi "},
			item{typ: itemKey, pos: 5, val: "@r"},
			item{typ: itemText, pos: 7, val: ", "},
			item{typ: itemKey, pos: 9, val: "@r"},
			item{typ: itemText, pos: 11, val: ", "},
			item{typ: itemSpecial, pos: 13, val: "$i12"},
			item{typ: itemNewLine, pos: 17, val: "\n"},
			item{typ: itemText, pos: 18, val: " lb "},
			item{typ: itemKey, pos: 22, val: "@r"},
			item{typ: itemText, pos: 24, val: ", "},
			item{typ: itemSpecial, pos: 26, val: "$i12"},
			item{typ: itemText, pos: 30, val: "("},
			item{typ: itemKey, pos: 31, val: "@r"},
			item{typ: itemText, pos: 33, val: ")"},
			item{typ: itemEOF, pos: 34, val: ""},
		}
		var actual []item
		for i := range l.items {
			actual = append(actual, i)
		}
		Equal(t, expected, actual)
	}
	{
		l := lex("$a")
		Equal(t, itemError, l.nextItem().typ)
	}
	{
		l := lex("$i a")
		Equal(t, itemError, l.nextItem().typ)
	}
	{
		l := lex("@\n")
		Equal(t, itemError, l.nextItem().typ)
	}
}
