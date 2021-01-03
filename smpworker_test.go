package main

import "testing"

func BenchmarkSMP(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		start("data")
	}
}
