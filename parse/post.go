package parse

import (
	"bytes"
	"strconv"
	"math/rand"
)

func PostProcess(s string, r *rand.Rand) string {
	l := lex(s)

	return replaceLabels(l, r)
}

func replaceLabels(l *lexer, r *rand.Rand) string {
	var buf bytes.Buffer
	var labelCounter uint
	var remLabels []uint

	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemText:
			buf.WriteString(i.val)
		case itemLabel:
			buf.WriteString("label")
			buf.WriteString(strconv.Itoa(int(labelCounter)))
			remLabels = append(remLabels, labelCounter)
			labelCounter++
		case itemNewLine:
			buf.WriteString("\n")

			// randomly put a label here, if any
			if len(remLabels) > 0 && r.Intn(8) == 0 {
				buf.WriteString("label")
				buf.WriteString(strconv.Itoa(int(remLabels[0])))
				buf.WriteString(":\n")
				remLabels = remLabels[1:]
			}
		}
	}

	for _, l := range remLabels {
		buf.WriteString("label")
		buf.WriteString(strconv.Itoa(int(l)))
		buf.WriteString(":\n")
	}

	return buf.String()
}
