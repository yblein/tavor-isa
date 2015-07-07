package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

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
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: <ISA configuration file> [fuzzing strategy]")
		printStrategies()
		os.Exit(1)
	}

	file := os.Args[1]
	root, err := parse.Parse(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	strategyName := defaultStrategyName
	if len(os.Args) >= 3 {
		strategyName = os.Args[2]
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

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	strat, err := strategy.New(strategyName, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		printStrategies()
		os.Exit(4)
	}

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
