package allfields

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"

	"github.com/upsun/vinego/src/utils"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:      "allfields",
		Doc:       "_",
		FactTypes: []analysis.Fact{new(utils.ChecksFact)},
		Run: func(p *analysis.Pass) (interface{}, error) {
			utils.ScanTypeTags(p)
			for _, file := range p.Files {
				ast.Inspect(file, func(n ast.Node) bool {
					literal, isCompLiteral := n.(*ast.CompositeLit)
					if !isCompLiteral {
						return true
					}
					litType := p.TypesInfo.TypeOf(literal)
					if litType == nil {
						return true
					}
					litType1, isNamed := litType.(*types.Named)
					if !isNamed {
						return true
					}
					enabledChecks := new(utils.ChecksFact)
					p.ImportObjectFact(litType1.Obj(), enabledChecks)
					if len(*enabledChecks) == 0 {
						return true
					}

					if (*enabledChecks)["allfields"] {
						structType, isStruct := litType.Underlying().(*types.Struct)
						if !isStruct {
							p.Report(analysis.Diagnostic{
								Pos:      n.Pos(),
								Category: "error",
								Message:  "Type marked as allfields is not a struct",
							})
							return true
						}
						remaining := map[string]bool{}
						for i := 0; i < structType.NumFields(); i++ {
							if reflect.StructTag(structType.Tag(i)).Get("optional") == "1" {
								continue
							}
							remaining[structType.Field(i).Name()] = true
						}

						nonKv := false
						for _, e0 := range literal.Elts {
							switch e := e0.(type) {
							case *ast.KeyValueExpr:
								switch k := e.Key.(type) {
								case *ast.Ident:
									delete(remaining, k.Name)
								}
							default:
								// TODO handle keyless fields?
								nonKv = true
							}
						}

						if !nonKv && len(remaining) > 0 {
							niceRemaining := []string{}
							for k := range remaining {
								niceRemaining = append(niceRemaining, k)
							}
							p.Report(analysis.Diagnostic{
								Pos:      n.Pos(),
								Category: "error",
								Message:  fmt.Sprintf("Missing required fields in struct literal: %v", niceRemaining),
							})
						}
					}

					return true
				})
			}
			return nil, nil
		},
	}
}
