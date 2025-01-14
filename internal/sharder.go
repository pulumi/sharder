package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
)

type TestRun struct {
	Test    string
	Elapsed float64
}

func sortRuns(runs *[]TestRun) {
	slices.SortStableFunc(*runs, func(a, b TestRun) int {
		if a.Elapsed > b.Elapsed {
			return -1
		}
		if a.Elapsed < b.Elapsed {
			return 1
		}
		if a.Test > b.Test {
			return -1
		}
		if a.Test < b.Test {
			return 1
		}
		return 0
	})
}

func ProcessJSON(filepath string) ([]TestRun, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	records := []TestRun{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		var record struct {
			Action  string  `json:"Action"`
			Test    *string `json:"Test"`
			Elapsed float64 `json:"Elapsed"`
		}
		// Parse the JSON line into a Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		if (record.Action == "pass" || record.Action == "fail") && record.Test != nil {
			records = append(records, TestRun{
				Test:    *record.Test,
				Elapsed: record.Elapsed,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	sortRuns(&records)
	return records, nil
}

func Aggregate(tests []TestRun) []TestRun {
	testMap := make(map[string]float64)

	for _, test := range tests {
		// strip the first part of the test name
		parts := strings.Split(test.Test, "/")
		if _, ok := testMap[parts[0]]; !ok {
			testMap[parts[0]] = 0
		}
		testMap[parts[0]] += test.Elapsed
	}

	res := make([]TestRun, 0)
	for test, elapsed := range testMap {
		res = append(res, TestRun{
			Test:    test,
			Elapsed: elapsed,
		})
	}

	sortRuns(&res)
	return res
}

type Bin struct {
	Tests []TestRun
	Total float64
}

func PackShards(tests []TestRun, total int, seed int64) []Bin {
	shortTestThreshold := 0.5
	bins := make([]Bin, total)

	longTests := []TestRun{}
	for _, test := range tests {
		if test.Elapsed > shortTestThreshold {
			longTests = append(longTests, test)
		}
	}

	// pack the tests into the bins according to their runtimes
	for _, test := range longTests {
		bins[0].Tests = append(bins[0].Tests, test)
		bins[0].Total += test.Elapsed

		// sort the bins by total runtime
		slices.SortStableFunc(bins, func(a, b Bin) int {
			if a.Total < b.Total {
				return -1
			}
			if a.Total > b.Total {
				return 1
			}
			return 0
		})
	}

	shortTests := []TestRun{}
	for _, test := range tests {
		if test.Elapsed <= shortTestThreshold {
			shortTests = append(shortTests, test)
		}
	}

	if seed != 0 {
		random := rand.New(rand.NewSource(seed)) //nolint:gosec // insecure random number generator is fine
		for i := range shortTests {
			j := random.Intn(i + 1)
			shortTests[i], shortTests[j] = shortTests[j], shortTests[i]
		}
	}

	for i, test := range shortTests {
		bin := bins[i%len(bins)]
		bin.Tests = append(bin.Tests, test)
		bin.Total += test.Elapsed
		bins[i%len(bins)] = bin
	}

	slices.Reverse(bins)
	return bins
}

func generateOutputLast(bins []Bin) string {
	names := make([]string, 0)
	// skip the last bin since it's the one we're running
	// we'll negate all the other bins
	for _, bin := range bins[0 : len(bins)-1] {
		for _, test := range bin.Tests {
			names = append(names, test.Test)
		}
	}

	pattern := ""
	for _, name := range names {
		pattern += fmt.Sprintf(`(^%s$)|`, name)
	}
	pattern = strings.TrimSuffix(pattern, "|")
	pattern = "\"" + pattern + "\""
	return "-skip " + pattern
}

func generateOutputNotLast(bins []Bin, shard int) string {
	names := make([]string, len(bins[shard].Tests))
	for i, test := range bins[shard].Tests {
		names[i] = test.Test
	}

	pattern := ""
	for _, name := range names {
		pattern += fmt.Sprintf(`(^%s$)|`, name)
	}
	pattern = strings.TrimSuffix(pattern, "|")
	pattern = "\"" + pattern + "\""
	return "-run " + pattern
}

func GenerateOutput(bins []Bin, index int) (string, error) {
	if index < 0 || index >= len(bins) {
		return "", fmt.Errorf("shard %d is out of bounds", index)
	}

	if index == len(bins)-1 {
		return generateOutputLast(bins), nil
	}

	return generateOutputNotLast(bins, index), nil
}
