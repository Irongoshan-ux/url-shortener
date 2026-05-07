package validation

import (
	"testing"
)

func BenchmarkParseHTTPURL(b *testing.B) {
	const raw = "https://example.com/some/deep/path/to/resource?x=1&y=2"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseHTTPURL(raw)
	}
}

func BenchmarkNormalizeBaseURL(b *testing.B) {
	const base = "https://shortener.example.com/api/"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeBaseURL(base)
	}
}
