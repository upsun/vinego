package capturederr

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/upsun/vinego/src/utils"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "capturederr",
		Doc:  "_",
		Run: func(p *analysis.Pass) (interface{}, error) {
			for _, file := range p.Files {
				// check:allfields
				type Layer struct {
					Declarations []token.Pos
				}
				for _, decl := range file.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						stack := []*Layer{
							{
								Declarations: []token.Pos{},
							},
						}

						check := func(ident *ast.Ident) {
							decl := p.TypesInfo.Uses[ident]
							if decl == nil {
								return
							}
							if decl.Type().String() != "error" {
								return
							}
							layer := utils.Last(stack)
							for _, v := range layer.Declarations {
								if decl.Pos() == v {
									return
								}
							}
							p.Report(analysis.Diagnostic{
								Pos:     ident.Pos(),
								Message: fmt.Sprintf("Assigning to captured err variable %s", decl.Name()),
							})
						}

						addFuncVarDecls := func(spec *ast.FuncType) {
							stackTop := stack[len(stack)-1]
							for _, f := range spec.Params.List {
								for _, name := range f.Names {
									stackTop.Declarations = append(
										stackTop.Declarations,
										name.Pos(),
									)
								}
							}
							for _, name := range utils.NamedReturns(spec) {
								stackTop.Declarations = append(
									stackTop.Declarations,
									name.Pos(),
								)
							}
						}

						addFuncVarDecls(d.Type)

						utils.Walk(
							d.Body,
							func(n0 ast.Node) (bool, bool) {
								stackTop := stack[len(stack)-1]
								switch n := n0.(type) {
								case *ast.DeclStmt:
									d := n.Decl.(*ast.GenDecl)
									for _, spec := range d.Specs {
										valSpec, isValSpec := spec.(*ast.ValueSpec)
										if !isValSpec {
											continue
										} else {
											for _, name := range valSpec.Names {
												stackTop.Declarations = append(
													stackTop.Declarations,
													name.Pos(),
												)
											}
										}
									}
								case *ast.AssignStmt:
									for _, l := range n.Lhs {
										ident, isIdent := l.(*ast.Ident)
										if !isIdent || ident.Name == "_" {
											continue
										}
										if n.Tok == token.ASSIGN {
											check(ident)
										} else {
											stackTop.Declarations = append(
												stackTop.Declarations,
												ident.Pos(),
											)
										}
									}
								case *ast.FuncLit:
									stack = append(stack, &Layer{
										Declarations: []token.Pos{},
									})
									addFuncVarDecls(n.Type)
									return true, true
								}
								return false, true
							},
							func(pushedLayer bool, n ast.Node) {
								if pushedLayer {
									stack = stack[:len(stack)-1]
								}
							},
						)
					}
				}
			}
			return nil, nil
		},
	}
}
