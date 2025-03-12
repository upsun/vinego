package explicitcast

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/upsun/vinego/src/utils"
)

func wantSet(names ...string) map[string]bool {
	keys := append(names, "interface{}", "any")
	out := map[string]bool{}
	for _, k := range keys {
		out[k] = true
		// TODO should get el if variadic rather than this hack
		out["[]"+k] = true
	}
	return out
}

func New() *analysis.Analyzer {
	wantString := wantSet("string")
	wantChar := wantSet("char", "rune", "byte")
	wantInts := wantSet(
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
	)
	wantFloats := wantSet(
		"float",
		"float32",
		"float64",
	)
	return &analysis.Analyzer{
		Name: "explicitcast",
		Doc:  "_",
		Run: func(p *analysis.Pass) (interface{}, error) {
			checkLit := func(p *analysis.Pass, t types.Type, e ast.Expr) {
				basicLit, isBasicLit := e.(*ast.BasicLit)
				if !isBasicLit {
					return
				}
				var want map[string]bool
				switch basicLit.Kind {
				case token.STRING:
					want = wantString
				case token.CHAR:
					want = wantChar
				case token.INT:
					want = wantInts
				case token.FLOAT, token.IMAG:
					want = wantFloats
				default:
					panic("ASSERTION! Unexpected literal kind")
				}
				if want[t.String()] {
					return
				}
				p.Report(analysis.Diagnostic{
					Pos:     e.Pos(),
					Message: fmt.Sprintf("Implicit literal cast of %s to %s", basicLit.Kind.String(), t.String()),
				})
			}

		NextFile:
			for _, file := range p.Files {
				for _, excl := range []string{
					"/usr",
					"/opt",
				} {
					if strings.HasPrefix(p.Fset.Position(file.Pos()).Filename, excl) {
						continue NextFile
					}
				}
				utils.WalkWithCrumbs(file, func(n0 ast.Node, crumbs []ast.Node) bool {
					switch n := n0.(type) {
					case *ast.CallExpr:
						funTypeObj := p.TypesInfo.Types[n.Fun]
						if !funTypeObj.IsValue() {
							// like `(func())(nil)` -- TypesInfo.TypeOf(fun) returns Sig same as a func obj, need to differentiate this way
							break
						}
						switch funType := funTypeObj.Type.(type) {
						case *types.Signature:
							if funType.Params().Len() > 1 && len(n.Args) == 1 {
								// function call multi-return forwarding - no implicit casts here
								break
							}
							for i, arg := range n.Args {
								var argType types.Type
								if funType.Variadic() && i >= funType.Params().Len()-1 {
									argType = funType.Params().At(funType.Params().Len() - 1).Type()
								} else {
									argType = funType.Params().At(i).Type()
								}
								checkLit(p, argType, arg)
							}
						case types.Type:
							// nop - ok
						}
					case *ast.AssignStmt:
						if len(n.Lhs) > 1 && len(n.Rhs) == 1 {
							// func return multi-assignment, no implicit casting here
							break
						}
						for i, dest := range n.Lhs {
							source := n.Rhs[i]
							destType := p.TypesInfo.TypeOf(dest)
							if destType != nil {
								checkLit(p, destType, source)
							}
						}
					case *ast.ReturnStmt:
						if len(n.Results) == 0 {
							break
						}
						var inFunc *types.Signature = nil
					FindFunc:
						for i := 0; i < len(crumbs); i++ {
							crumb := crumbs[len(crumbs)-1-i]
							switch f := crumb.(type) {
							case *ast.FuncDecl:
								inFunc = p.TypesInfo.TypeOf(f.Name).(*types.Signature)
								break FindFunc
							case *ast.FuncLit:
								inFunc = p.TypesInfo.TypeOf(f).(*types.Signature)
								break FindFunc
							}
						}
						if inFunc == nil {
							t := strings.Builder{}
							for i, c := range crumbs {
								t.WriteString(fmt.Sprintf("crumb %d %T %s\n", i, c, p.Fset.Position(c.Pos())))
							}
							panic("ASSERTION!" + p.String() + "\n" + t.String())
						}
						if inFunc.Results().Len() > 1 && len(n.Results) == 1 {
							// forward function call multi-return, no implicit casting here
							break
						}
						for i := 0; i < inFunc.Results().Len(); i++ {
							retType := inFunc.Results().At(i)
							checkLit(p, retType.Type(), n.Results[i])
						}
					}
					return true
				})
			}
			return nil, nil
		},
	}
}
