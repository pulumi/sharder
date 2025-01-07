package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/VenelinMartinov/sharder/internal"
)

func main() {
	outputFile := flag.String("output", "", "output file containing the test run times")
	seed := flag.Int64("seed", 0, "randomly shuffle tests using this seed")
	total := flag.Int("total", 1, "total number of shards")
	shard := flag.Int("shard", 0, "shard number")
	format := flag.String("format", "", "output format")

	flag.Parse()

	if *outputFile == "" {
		log.Fatalf("Error: output file is required")
	}

	result, err := internal.ProcessJSON(*outputFile)
	if err != nil {
		log.Fatalf("Error processing JSON: %v", err)
	}

	result = internal.Aggregate(result)

	shards := internal.PackShards(result, *total, *seed)

	pattern, err := internal.GenerateOutput(shards, *shard)
	if err != nil {
		log.Fatalf("Error generating output: %v", err)
	}

	if *format == "env" {
		fmt.Fprintf(os.Stdout, "%s=%s", "SHARD_CMD", pattern)
	} else {
		fmt.Fprintln(os.Stdout, pattern)
	}
}
