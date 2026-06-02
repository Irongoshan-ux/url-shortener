package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "reports log.Fatal and os.Exit outside func main of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			kind := forbiddenCall(sel)
			if kind == "" {
				return true
			}
			if allowedInMain(pass, call) {
				return true
			}
			switch kind {
			case "log.Fatal":
				pass.Reportf(call.Pos(), "log.Fatal outside main")
			case "os.Exit":
				pass.Reportf(call.Pos(), "os.Exit outside main")
			}
			return true
		})
	}
	return nil, nil
}

func forbiddenCall(sel *ast.SelectorExpr) string {
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	switch x.Name {
	case "log":
		if sel.Sel.Name == "Fatal" || sel.Sel.Name == "Fatalf" {
			return "log.Fatal"
		}
	case "os":
		if sel.Sel.Name == "Exit" {
			return "os.Exit"
		}
	}
	return ""
}

func allowedInMain(pass *analysis.Pass, n ast.Node) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}
	for _, f := range pass.Files {
		path, _ := astutil.PathEnclosingInterval(f, n.Pos(), n.End())
		for _, node := range path {
			fn, ok := node.(*ast.FuncDecl)
			if ok && fn.Name != nil && fn.Name.Name == "main" && fn.Recv == nil {
				return true
			}
		}
	}
	return false
}
