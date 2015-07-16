package main

import (
	"strings"
	"testing"

	. "github.com/zimmski/tavor/test/assert"

	"github.com/zimmski/tavor/fuzz/strategy"
	"github.com/zimmski/tavor/log"
	"github.com/zimmski/tavor/parser"
	"github.com/zimmski/tavor/test"
	//"github.com/zimmski/tavor/token"
)

func TestTokenCoverageToBeStrategy(t *testing.T) {
	var strat *strategy.Strategy

	Implements(t, strat, &TokenCoverage{})
}

func TestTokenCoverage(t *testing.T) {
	log.LevelDebug()

	validateTavorTokenCoverage(
		t,
		`
		START = (0 | 1) 2 (3 | 4)
		`,
		[]string{
			"023",
			"124",
		},
	)

	// enable an option only if its token is not already fully covered
	validateTavorTokenCoverage(
		t,
		`
		START = (1 | 2) ?(3)
		`,
		[]string{
			"13",
			"2",
		},
	)

	validateTavorTokenCoverage(
		t,
		`
		$Num Int
		START = Num | [\d]
		`,
		[]string{
			"0",
			"0",
		},
	)

	validateTavorTokenCoverage(
		t,
		`
		START = (0 (1 | 2 | 3 ?("a")) 4 (5 | 6)) | "b"
		`,
		[]string{
			"03a45",
			"0146",
			"0245",
			"b",
		},
	)

	// Test that repetitions are used to maximize the coverage of ONE test
	// and that they are not repeated unnecessarily
	// e.g., 2 repetitions out of 3 are enough here
	validateTavorTokenCoverage(
		t,
		`
		START = +1,3((0 | 1) 2 (3 | 4) " ")
		`,
		[]string{
			"023 124 ",
		},
	)

	// Test that repetitions are re-used properly accross multiple tests
	// when necessary
	validateTavorTokenCoverage(
		t,
		`
		START = +1,2((0 | 1 | 2) " ")
		`,
		[]string{
			"0 1 ",
			"2 0 ",
		},
	)

}

func validateTavorTokenCoverage(t *testing.T, format string, expect []string) {
	r := test.NewRandTest(1)

	o, err := parser.ParseTavor(strings.NewReader(format))
	Nil(t, err)

	s := NewTokenCoverage(o)

	ch, err := s.Fuzz(r)
	Nil(t, err)

	var got []string

	for i := range ch {
		got = append(got, o.String())

		ch <- i
	}

	Equal(t, expect, got)
}

/*
func TestTokenCoverageLoopDetection(t *testing.T) {
	testStrategyLoopDetection(t, func(root token.Token) strategy.Strategy {
		return NewTokenCoverage(root)
	})
}
*/
