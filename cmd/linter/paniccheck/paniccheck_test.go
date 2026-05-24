package paniccheck_test

import (
	"testing"

	"github.com/Irongoshan-ux/url-shortener/cmd/linter/paniccheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, paniccheck.Analyzer, "check")
}
