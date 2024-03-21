package utils

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type ChecksFact map[string]bool

func (*ChecksFact) AFact() {}

var checkRegexp = regexp.MustCompile("check:([a-z]+)")

func ScanTypeTags(p *analysis.Pass) {
	for _, file := range p.Files {
		ast.Inspect(file, func(n0 ast.Node) bool {
			decl, isGenDecl := n0.(*ast.GenDecl)
			if !isGenDecl {
				return true
			}
			if decl.Doc == nil {
				return true
			}
			enabledChecks := &ChecksFact{}
			for _, line := range strings.Split(decl.Doc.Text(), "\n") {
				matches := checkRegexp.FindStringSubmatch(line)
				if len(matches) >= 2 {
					(*enabledChecks)[matches[1]] = true
				}
			}
			if len(*enabledChecks) == 0 {
				return true
			}
			for _, declElem0 := range decl.Specs {
				typeSpec, isTypeSpec := declElem0.(*ast.TypeSpec)
				if !isTypeSpec {
					continue
				}
				declDef := p.TypesInfo.Defs[typeSpec.Name]
				p.ExportObjectFact(declDef, enabledChecks)
			}
			return true
		})
	}
}

func GetTypeTags(p *analysis.Pass, o types.Object) ChecksFact {
	enabledChecks := new(ChecksFact)
	p.ImportObjectFact(o, enabledChecks)
	return *enabledChecks
}
