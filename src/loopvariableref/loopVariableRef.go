package loopvariableref

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/platformsh/vinego/utils"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "loopvariableref",
		Doc:  "_",
		Run: func(p *analysis.Pass) (interface{}, error) {
			for _, file := range p.Files {
				// check:allfields
				type Layer struct {
					LayerType             string // func,return
					SensitiveDeclarations []token.Pos
				}
				seenUses := map[token.Pos]bool{}
				stack := []*Layer{
					{
						SensitiveDeclarations: []token.Pos{},
					},
				}

				check := func(initialRisky bool, context string, ident *ast.Ident) {
					decl := p.TypesInfo.Uses[ident]
					if decl == nil {
						return
					}
					risky := initialRisky
				StackLoop:
					for j := 0; j < len(stack); j += 1 {
						at := len(stack) - j - 1
						layer := stack[at]
						for _, v := range layer.SensitiveDeclarations {
							if decl.Pos() == v {
								if risky {
									p.Report(analysis.Diagnostic{
										Pos:     ident.Pos(),
										Message: fmt.Sprintf("Risky %s loop variable %s", context, ident.Name),
									})
								}
								seenUses[ident.Pos()] = true
								break StackLoop
							}
						}
						switch layer.LayerType {
						case "func":
							risky = true
						case "return":
							risky = false
						}
					}
				}

				utils.Walk(
					file,
					func(n0 ast.Node) (bool, bool) {
						stackTop := stack[len(stack)-1]
						switch n := n0.(type) {
						case *ast.RangeStmt:
							// When we find a for each loop, add the loop variables to our watches list
							if n.Tok != token.DEFINE {
								break
							}
							stackTop.SensitiveDeclarations = append(
								stackTop.SensitiveDeclarations,
								n.Key.Pos(),
							)
							if n.Value != nil {
								stackTop.SensitiveDeclarations = append(
									stackTop.SensitiveDeclarations,
									n.Value.Pos(),
								)
							}
						case *ast.UnaryExpr:
							// If taking a reference to something, check that it's not a loop variable
							if n.Op != token.AND {
								break
							}
							switch x := n.X.(type) {
							case *ast.Ident:
								check(true, "reference to", x)
							case *ast.SelectorExpr:
								ident, isIdent := x.X.(*ast.Ident)
								if !isIdent {
									break
								}
								check(true, "reference to child of", ident)
							}
						case *ast.FuncLit:
							// Add a new layer for each function literal to demarcate all subsequent
							// uses of previously defined loop variables as unsafe
							stack = append(stack, &Layer{
								LayerType:             "func",
								SensitiveDeclarations: []token.Pos{},
							})
							return true, true
						case *ast.ReturnStmt:
							// Return statements negate the sensitivity of all higher loops in the function since they won't iterate any further
							stack = append(stack, &Layer{
								LayerType:             "return",
								SensitiveDeclarations: []token.Pos{},
							})
							return true, true
						case *ast.Ident:
							// Normal (non-reference) usage of a variable - check if it's a capture
							// (defined outside of current function literal)
							if seenUses[n.Pos()] {
								break
							}
							decl := p.TypesInfo.Uses[n]
							if decl == nil {
								break
							}
							check(false, "capture of", n)
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
			return nil, nil
		},
	}
}
