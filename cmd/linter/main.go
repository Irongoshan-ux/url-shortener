package main

import (
	"github.com/Irongoshan-ux/url-shortener/cmd/linter/osexit"
	"github.com/Irongoshan-ux/url-shortener/cmd/linter/paniccheck"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		paniccheck.Analyzer,
		osexit.Analyzer,
	)
}
