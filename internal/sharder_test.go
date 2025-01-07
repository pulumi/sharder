package internal

import (
	"os"
	"reflect"
	"testing"
)

func TestShard(t *testing.T) {
	t.Parallel()
	jsonInput := []byte(`{"Action": "pass", "Test": "test1", "Elapsed": 1.23}
		{"Action": "fail", "Test": "test2", "Elapsed": 2.34}
		{"Action": "pass", "Test": "test3", "Elapsed": 0.56}
		{"Action": "pass", "Test": null, "Elapsed": 0.78}`)

	tmpfile, err := os.CreateTemp("", "test.json")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err = tmpfile.Write(jsonInput); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatalf("Error closing temp file: %v", err)
	}

	result, err := ProcessJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error processing JSON: %v", err)
	}

	expected := []TestRun{
		{Test: "test2", Elapsed: 2.34},
		{Test: "test1", Elapsed: 1.23},
		{Test: "test3", Elapsed: 0.56},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPackShards(t *testing.T) {
	t.Parallel()

	tests := []TestRun{
		{Test: "long_test1", Elapsed: 2.5},
		{Test: "long_test2", Elapsed: 1.5},
		{Test: "long_test3", Elapsed: 1.5},
		{Test: "long_test4", Elapsed: 1.5},
		{Test: "short_test1", Elapsed: 0.4},
		{Test: "short_test2", Elapsed: 0.3},
		{Test: "short_test3", Elapsed: 0.2},
		{Test: "short_test4", Elapsed: 0.1},
	}

	bins := PackShards(tests, 2, 0)

	// Verify we got the expected number of bins
	if len(bins) != 2 {
		t.Errorf("Expected 2 bins, got %d", len(bins))
	}

	expected := []Bin{
		{
			Tests: []TestRun{
				{
					Test:    "long_test1",
					Elapsed: 2.5,
				},
				{
					Test:    "long_test4",
					Elapsed: 1.5,
				},
				{
					Test:    "short_test2",
					Elapsed: 0.3,
				},
				{
					Test:    "short_test4",
					Elapsed: 0.1,
				},
			},
			Total: 4.3999999999999995,
		},
		{
			Tests: []TestRun{
				{
					Test:    "long_test2",
					Elapsed: 1.5,
				},
				{
					Test:    "long_test3",
					Elapsed: 1.5,
				},
				{
					Test:    "short_test1",
					Elapsed: 0.4,
				},
				{
					Test:    "short_test3",
					Elapsed: 0.2,
				},
			},
			Total: 3.6,
		},
	}

	if !reflect.DeepEqual(bins, expected) {
		t.Errorf("Expected %v, got %v", expected, bins)
	}
}

func TestAggregate(t *testing.T) {
	t.Parallel()

	tests := []TestRun{
		{Test: "test3", Elapsed: 1.0},
		{Test: "test1", Elapsed: 3.0},
		{Test: "test1", Elapsed: 1.0},
		{Test: "test2", Elapsed: 2.0},
	}

	result := Aggregate(tests)

	expected := []TestRun{
		{Test: "test1", Elapsed: 4.0},
		{Test: "test2", Elapsed: 2.0},
		{Test: "test3", Elapsed: 1.0},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestGenerateOutput(t *testing.T) {
	t.Parallel()

	bins := []Bin{
		{
			Tests: []TestRun{
				{Test: "long_test2", Elapsed: 1.5},
				{Test: "long_test3", Elapsed: 1.5},
				{Test: "short_test1", Elapsed: 0.4},
				{Test: "short_test3", Elapsed: 0.2},
			},
		},
		{
			Tests: []TestRun{
				{Test: "long_test1", Elapsed: 2.5},
				{Test: "long_test4", Elapsed: 1.5},
				{Test: "short_test2", Elapsed: 0.3},
				{Test: "short_test4", Elapsed: 0.1},
			},
		},
	}

	pattern, err := GenerateOutput(bins, 0)
	if err != nil {
		t.Fatalf("Error generating output: %v", err)
	}

	expected := `-run "(^long_test2$)|(^long_test3$)|(^short_test1$)|(^short_test3$)"`
	if pattern != expected {
		t.Errorf("Expected %v, got %v", expected, pattern)
	}
}
