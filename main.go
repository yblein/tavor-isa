package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
	"flag"

	"github.com/yblein/tavor-isa/parse"

	"github.com/zimmski/tavor"
	"github.com/zimmski/tavor/fuzz/filter"
	"github.com/zimmski/tavor/fuzz/strategy"
	//"github.com/zimmski/tavor/graph"
	//"github.com/zimmski/tavor/log"
)

const (
	DefaultStrategyName = "random"
	TemporaryFile = "TAVOR_FUZZ_FILE"
)

func printStrategies() {
	fmt.Fprintln(os.Stderr, "Available strategies:")
	for _, s := range strategy.List() {
		if s == DefaultStrategyName {
			fmt.Fprintln(os.Stderr, "-", s, "(default)")
		} else {
			fmt.Fprintln(os.Stderr, "-", s)
		}
	}
}

func main() {
	seed := flag.Int64("seed", -1, "seed for randomness")
	strategyName := flag.String("strategy", DefaultStrategyName, "fuzzing strategy")
	execFlag := flag.String("exec", "", "execute this script; tests are produced into a temporary file which is defined using the environment variable TAVOR_FUZZ_FILE")

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

	outputFilename := ""
	if *execFlag != "" {
		outputFilename = os.Getenv(TemporaryFile)
		if outputFilename == "" {
			fmt.Fprintln(os.Stderr, "error: the environment variable", TemporaryFile, "must be set when using the `execFlag` option")
			os.Exit(2)
		}
	}

	tavor.MaxRepeat = 100

	file := flag.Arg(0)
	root, err := parse.Parse(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	//graph.WriteDot(root, os.Stdout)

	//log.LevelDebug()

	var filters = []filter.Filter{
		filter.NewPositiveBoundaryValueAnalysisFilter(),
	}

	root, err = filter.ApplyFilters(filters, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	strat, err := strategy.New(*strategyName, root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		printStrategies()
		os.Exit(5)
	}

	if *seed < 0 {
		*seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(*seed))

	continueFuzzing, err := strat.Fuzz(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(6)
	}

	for i := range continueFuzzing {
		if *execFlag == "" {
			fmt.Println(root.String())
		} else {
			outputFile, err := os.Create(outputFilename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(7)
			}

			outputFile.WriteString(root.String())

			// exec the command
			cmd := exec.Command(*execFlag)
			err = cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error when executing the command:", err)
				os.Exit(8)
			}
		}

		continueFuzzing <- i
	}
}
