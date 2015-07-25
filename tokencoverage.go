package main

import (
	"fmt"

	"github.com/zimmski/tavor/fuzz/strategy"
	"github.com/zimmski/tavor/log"
	"github.com/zimmski/tavor/rand"
	"github.com/zimmski/tavor/token"
	"github.com/zimmski/tavor/token/constraints"
	"github.com/zimmski/tavor/token/lists"
	"github.com/zimmski/tavor/token/primitives"
)

// TokenCoverage implements a fuzzing strategy that produces a covering of all the tokens of the token graph.
// The strategy produces a set of tests such that all the tokens of the token graphs have been covered by at least one test.
type TokenCoverage struct {
	root         token.Token
	covered      map[token.Token]struct{}
	repeatCovers map[*lists.Repeat]map[token.Token]struct{}
	path         []uint
}

// NewTokenCoverage returns a new instance of the token coverage fuzzing strategy
func NewTokenCoverage(tok token.Token) *TokenCoverage {
	return &TokenCoverage{
		root:         tok,
		covered:      make(map[token.Token]struct{}),
		repeatCovers: make(map[*lists.Repeat]map[token.Token]struct{}),
	}
}

func init() {
	strategy.Register("TokenCoverage", func(tok token.Token) strategy.Strategy {
		return NewTokenCoverage(tok)
	})
}

// Fuzz starts the first iteration of the fuzzing strategy returning a channel which controls the iteration flow.
// The channel returns a value if the iteration is complete and waits with calculating the next iteration until a value is put in. The channel is automatically closed when there are no more iterations. The error return argument is not nil if an error occurs during the setup of the fuzzing strategy.
func (s *TokenCoverage) Fuzz(r rand.Rand) (chan struct{}, error) {
	if token.LoopExists(s.root) {
		return nil, &strategy.Error{
			Message: "found endless loop in graph. Cannot proceed.",
			Type:    strategy.ErrorEndlessLoopDetected,
		}
	}

	continueFuzzing := make(chan struct{})

	go func() {
		for {
			nbUncovered, path := s.bestUncoveredPath(s.root)

			if nbUncovered == 0 {
				// all tokens have been covered, stop fuzzing
				close(continueFuzzing)
				return
			}

			s.path = path
			s.setPath(s.root)

			// done with the last fuzzing step
			continueFuzzing <- struct{}{}

			// wait until we are allowed to continue
			if _, ok := <-continueFuzzing; !ok {
				log.Debug("fuzzing channel closed from outside")
				return
			}

			token.ResetCombinedScope(s.root)
			token.ResetResetTokens(s.root)
			token.ResetCombinedScope(s.root)
		}
	}()

	return continueFuzzing, nil
}

func (s *TokenCoverage) bestUncoveredPath(tok token.Token) (uint, []uint) {
	switch t := tok.(type) {
	case *constraints.Optional:
		// activate the option only if there is uncovered tokens in it
		nbUncovered, path := s.bestUncoveredPath(t.InternalGet())
		bestPermutation := uint(1)

		if nbUncovered >= 1 {
			bestPermutation = 2
		}

		return nbUncovered, append([]uint{bestPermutation}, path...)

	case *lists.One:
		// find the permutation leading to the highest number of uncovered tokens
		l := t.InternalLen()

		var bestPermutation uint
		var bestNbUncovered uint
		var bestPath []uint

		for i := 0; i < l; i++ {
			c, _ := t.InternalGet(i)
			nbUncovered, path := s.bestUncoveredPath(c)

			if nbUncovered > bestNbUncovered || bestPermutation == 0 {
				bestPermutation = uint(i + 1)
				bestNbUncovered = nbUncovered
				bestPath = path
			}
		}

		return bestNbUncovered, append([]uint{bestPermutation}, bestPath...)

	case *lists.All:
		// go through all the tokens of the list
		var nbUncoveredTotal uint
		completePath := []uint{1}

		for i := 0; i < t.InternalLen(); i++ {
			c, _ := t.InternalGet(i)
			nbUncovered, path := s.bestUncoveredPath(c)
			nbUncoveredTotal += nbUncovered
			completePath = append(completePath, path...)
		}

		return nbUncoveredTotal, completePath

	case *lists.Repeat:
		// repeat as many times as necessary to cover all the underlying tokens
		c, _ := t.InternalGet(0)
		coveredBakup := s.covered

		if covered, ok := s.repeatCovers[t]; ok {
			s.covered = covered
		} else {
			s.covered = make(map[token.Token]struct{})
			s.repeatCovers[t] = s.covered
		}

		var nbUncoveredTotal uint
		completePath := []uint{t.Permutations()}

		for i := int64(0); i < t.To(); i++ {
			nbUncovered, path := s.bestUncoveredPath(c)

			if i > t.From() && nbUncovered == 0 {
				completePath[0] = uint(i - t.From() + 1)
				break
			}

			s.path = path
			s.setPath(c)

			nbUncoveredTotal += nbUncovered
			completePath = append(completePath, path...)
		}

		s.covered = coveredBakup

		return nbUncoveredTotal, completePath

	case *primitives.CharacterClass, *primitives.ConstantString, *primitives.ConstantInt, *primitives.RangeInt:
		var nbUncovered uint
		if _, covered := s.covered[tok]; !covered {
			nbUncovered = 1
		}
		return nbUncovered, []uint{1}

	case token.Forward:
		nbUncovered, path := s.bestUncoveredPath(t.InternalGet())
		return nbUncovered, append([]uint{1}, path...)

	default:
		panic(fmt.Errorf("bestUncoveredPath not implemented for %#v", t))
	}
}

func (s *TokenCoverage) setPath(tok token.Token) {
	_ = tok.Permutation(s.path[0])
	s.path = s.path[1:]
	s.covered[tok] = struct{}{}

	switch t := tok.(type) {
	case token.List:
		for i := 0; i < t.Len(); i++ {
			c, _ := t.Get(i)
			s.setPath(c)
		}
	case token.Forward:
		s.setPath(t.InternalGet())
	default:
	}
}
