package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
	"flag"

	"github.com/yblein/tavor-isa/parse"

	"github.com/zimmski/tavor/fuzz/filter"
	"github.com/zimmski/tavor/fuzz/strategy"
	//"github.com/zimmski/tavor/graph"
	//"github.com/zimmski/tavor/log"
)

const defaultStrategyName = "random"

func printStrategies() {
	fmt.Fprintln(os.Stderr, "Available strategies:")
	for _, s := range strategy.List() {
		if s == defaultStrategyName {
			fmt.Fprintln(os.Stderr, "-", s, "(default)")
		} else {
			fmt.Fprintln(os.Stderr, "-", s)
		}
	}
}

func main() {
	seed := flag.Int64("seed", -1, "seed for randomness")
	strategyName := flag.String("strategy", defaultStrategyName, "fuzzing strategy")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <ISA configuration file>\nOptionnal flags:\n", os.Args[0])
		flag.PrintDefaults()
		printStrategies()
	}

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	file := flag.Arg(0)
	root, err := parse.Parse(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	//graph.WriteDot(root, os.Stdout)

	//log.LevelDebug()

	var filters = []filter.Filter{
		filter.NewPositiveBoundaryValueAnalysisFilter(),
	}

	root, err = filter.ApplyFilters(filters, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	strat, err := strategy.New(*strategyName, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		printStrategies()
		os.Exit(4)
	}

	if *seed < 0 {
		*seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(*seed))

	continueFuzzing, err := strat.Fuzz(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}

	for i := range continueFuzzing {
		fmt.Println(root.String())

		continueFuzzing <- i
	}
}
