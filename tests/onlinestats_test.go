package main

import (
	"testing"
)

const (
	targetFile = "./data/XXX.csv"
)

func BenchmarkDataNaive(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()

	for nthTest := 0; nthTest < b.N; nthTest++ {
		// Continue
	}
}
