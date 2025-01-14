package main

import (
	"testing"
)

func TestIntegration(t *testing.T) {
	outs := []string{}

	for i := 0; i < 3; i++ {
		out, err := run(runInput{
			OutputFile: "testdata/test-timings.jsonl",
			Seed:       0,
			Total:      3,
			Index:      i,
			Format:     "",
		})
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		outs = append(outs, out)
	}

	expected := []string{
		"-run \"(^test3$)|(^short2$)\"",
		"-run \"(^test2$)|(^short3$)|(^short1$)\"",
		"-skip \"(^test3$)|(^short2$)|(^test2$)|(^short3$)|(^short1$)\"",
	}

	for i, out := range outs {
		if out != expected[i] {
			t.Fatalf("Expected %v %v, got %v", i, expected[i], out)
		}
	}
}
