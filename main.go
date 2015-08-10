package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/yblein/tavor-isa/parse"

	"github.com/zimmski/tavor"
	"github.com/zimmski/tavor/fuzz/filter"
	"github.com/zimmski/tavor/fuzz/strategy"
	//"github.com/zimmski/tavor/graph"
	//"github.com/zimmski/tavor/log"
)

const (
	defaultStrategyName    = "TokenCoverage"
	defaultMaxInstructions = 3000
)

func printStrategies() {
	fmt.Fprintln(os.Stderr, "\nAvailable fuzzing strategies:")
	for _, s := range strategy.List() {
		if s == defaultStrategyName {
			fmt.Fprintln(os.Stderr, "-", s, "(default)")
		} else {
			fmt.Fprintln(os.Stderr, "-", s)
		}
	}
}

func main() {
	flagSet := flag.NewFlagSet("flags", flag.ExitOnError)

	seed := flagSet.Int64("seed", -1, "seed for randomness")
	strategyName := flagSet.String("strategy", defaultStrategyName, "fuzzing strategy")
	execFlag := flagSet.String("exec", "", "execute this script with the test file as argument")
	maxInstructions := flagSet.Int("max-instructions", defaultMaxInstructions, "maximum number of instructions per test program")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <ISA configuration file>\n\nOptionnal flags:\n", os.Args[0])
		flagSet.PrintDefaults()
		printStrategies()
	}

	_ = flagSet.Parse(os.Args[1:])

	if len(flagSet.Args()) < 1 {
		flagSet.Usage()
		os.Exit(1)
	}

	var outputFile *os.File
	if *execFlag != "" {
		var err error
		outputFile, err = ioutil.TempFile(os.TempDir(), "tavor-isa")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		defer func() {
			_ = outputFile.Close()
			_ = os.Remove(outputFile.Name())
		}()
	}

	tavor.MaxRepeat = *maxInstructions

	file := flagSet.Arg(0)
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
		s := parse.PostProcess(root.String(), r)

		if *execFlag == "" {
			fmt.Println(s)
		} else {
			_, _ = outputFile.Seek(0, 0)
			n, _ := outputFile.WriteString(s)
			_ = outputFile.Truncate(int64(n))

			cmd := exec.Command(*execFlag, outputFile.Name())
			err = cmd.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error when executing the command `%s %s`: %s\n", *execFlag, outputFile.Name(), err)
				os.Exit(7)
			}
		}

		continueFuzzing <- i
	}
}
