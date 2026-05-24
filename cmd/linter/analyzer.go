package main

import (
	"github.com/Irongoshan-ux/url-shortener/cmd/linter/osexit"
	"github.com/Irongoshan-ux/url-shortener/cmd/linter/paniccheck"
	"golang.org/x/tools/go/analysis"
)

// analyzer runs all project lint checks in one pass for singlechecker.
var analyzer = &analysis.Analyzer{
	Name: "linter",
	Doc:  "checks for panic, log.Fatal, and os.Exit misuse",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if _, err := paniccheck.Analyzer.Run(pass); err != nil {
		return nil, err
	}
	if _, err := osexit.Analyzer.Run(pass); err != nil {
		return nil, err
	}
	return nil, nil
}
