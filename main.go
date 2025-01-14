package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pulumi/sharder/internal"
)

func main() {
	outputFile := flag.String("output", "", "output file containing the test run times")
	seed := flag.Int64("seed", 0, "randomly shuffle tests using this seed")
	total := flag.Int("total", 1, "total number of shards")
	index := flag.Int("index", 0, "shard index")
	format := flag.String("format", "", "output format")

	flag.Parse()

	out, err := run(runInput{
		OutputFile: *outputFile,
		Seed:       *seed,
		Total:      *total,
		Index:      *index,
		Format:     *format,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stdout, out)
}

type runInput struct {
	OutputFile string
	Seed       int64
	Total      int
	Index      int
	Format     string
}

func run(input runInput) (string, error) {
	if input.OutputFile == "" {
		return "", fmt.Errorf("error: output file is required")
	}

	result, err := internal.ProcessJSON(input.OutputFile)
	if err != nil {
		return "", fmt.Errorf("error processing JSON: %w", err)
	}

	result = internal.Aggregate(result)

	shards := internal.PackShards(result, input.Total, input.Seed)

	pattern, err := internal.GenerateOutput(shards, input.Index)
	if err != nil {
		return "", fmt.Errorf("error generating output: %w", err)
	}

	if input.Format == "make" {
		pattern = strings.ReplaceAll(pattern, "$", "\\$$")
		return pattern, nil
	}

	return pattern, nil
}
