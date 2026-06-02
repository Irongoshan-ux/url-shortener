package osexit_test

import (
	"testing"

	"github.com/Irongoshan-ux/url-shortener/cmd/linter/osexit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, osexit.Analyzer, "mainpkg", "lib")
}
