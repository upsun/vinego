package varinit

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/upsun/vinego/src/semivendor/gocfg/cfg"
	"github.com/upsun/vinego/src/semivendor/gocfg/ctrlflow"
	"github.com/upsun/vinego/src/utils"
)

type BranchId token.Pos
type VarId token.Pos

type DeclBranch struct {
	Comment string
}

type Decl struct {
	Name string
	// If this decl initialization has changed
	Changed       bool
	Uninitialized map[BranchId]DeclBranch
}

// check:allfields
type Scope struct {
	Location      BranchId
	Comment       string
	Uninitialized map[VarId]*Decl
}

// check:allfields
type Context struct {
	p        *analysis.Pass
	cfgs     *ctrlflow.CFGs
	scope    *Scope
	reported map[VarId]bool
}

func DeclIdForUse(p *analysis.Pass, ident *ast.Ident) VarId {
	obj := p.TypesInfo.Uses[ident]
	if obj == nil {
		return VarId(0)
	}
	id := VarId(obj.Pos())
	return id
}

func (s *Scope) GetUninitialized(p *analysis.Pass, id VarId) *Decl {
	return s.Uninitialized[id]
}

func (s *Scope) NewDecl(p *analysis.Pass, ident *ast.Ident) {
	obj := p.TypesInfo.Defs[ident]
	if obj == nil {
		return
	}
	id := VarId(obj.Pos())
	s.Uninitialized[id] = &Decl{
		Name:          ident.Name,
		Changed:       false,
		Uninitialized: map[BranchId]DeclBranch{s.Location: {Comment: s.Comment}},
	}
}

func (s *Scope) MarkInitialized(p *analysis.Pass, ident *ast.Ident) {
	obj := p.TypesInfo.Uses[ident]
	if obj == nil {
		return
	}
	id := VarId(obj.Pos())
	decl, hasDecl := s.Uninitialized[id]
	if hasDecl {
		decl.Changed = true
		decl.Uninitialized = nil
	}
}

// Skips root element -- this is only for when other specific evaluations fall through (i.e. root element has no useful info)
func Recurse(c *Context, n0 ast.Node) {
	ast.Inspect(n0, func(n ast.Node) bool {
		if n == n0 {
			return true
		}
		switch n1 := n.(type) {
		case ast.Stmt:
			EvalStmt(c, n1)
			return false
		case ast.Expr:
			EvalExpr(c, n1)
			return false
		default:
			return true
		}
	})
}

func EvalVarDecl(c *Context, valSpec *ast.ValueSpec) {
	// go syntax invariant
	// len(vals) == 0 or len(vals) == len(names)
	if len(valSpec.Values) > 0 {
		for _, val := range valSpec.Values {
			Recurse(c, val)
		}
	} else {
		for _, name := range valSpec.Names {
			c.scope.NewDecl(c.p, name)
		}
	}
}

func EvalVarDeclBlock(c *Context, d *ast.GenDecl) {
	if d.Tok == token.CONST {
		return
	}
	for _, spec := range d.Specs {
		valSpec, isValSpec := spec.(*ast.ValueSpec)
		if !isValSpec {
			Recurse(c, d)
		} else {
			EvalVarDecl(c, valSpec)
		}
	}
}

func CheckUseByDecl(c *Context, id VarId, name string, reportPos token.Pos) {
	uninit := c.scope.GetUninitialized(c.p, id)
	if uninit == nil || len(uninit.Uninitialized) == 0 {
		return
	}
	branchStrings := []string{}
	for branch := range uninit.Uninitialized {
		utils.Append(&branchStrings, " - "+c.p.Fset.Position(token.Pos(branch)).String())
	}
	c.p.Report(analysis.Diagnostic{
		Pos:     reportPos,
		Message: fmt.Sprintf("`%s` hasn't been initialized in the following branches:\n%s\n", uninit.Name, strings.Join(branchStrings, "\n")),
	})
	c.reported[id] = true
}

func CheckUse(c *Context, e *ast.Ident) {
	declId := DeclIdForUse(c.p, e)
	if c.reported[declId] {
		return
	}
	CheckUseByDecl(c, declId, e.Name, e.Pos())
}

func EvalExpr(c *Context, n ast.Expr) {
	switch e := n.(type) {
	case *ast.CallExpr:
		switch f := e.Fun.(type) {
		case *ast.FuncLit:
			for _, arg := range e.Args {
				EvalExpr(c, arg)
			}
			resScope := EvalFunc(c.p, c.cfgs, c.cfgs.FuncLit(f), f.Type, []*Scope{c.scope}, c.reported)
			c.scope.Uninitialized = resScope.Uninitialized
		default:
			Recurse(c, e)
		}
	case *ast.UnaryExpr:
		{
			ident, isIdent := e.X.(*ast.Ident)
			if !isIdent {
				goto NotAddr
			}
			if e.Op != token.AND {
				goto NotAddr
			}
			c.scope.MarkInitialized(c.p, ident)
			return
		}
	NotAddr:
		Recurse(c, e)
	case *ast.Ident:
		CheckUse(c, e)
	default:
		Recurse(c, e)
	}
}

func EvalStmt(c *Context, n ast.Stmt) {
	switch s := n.(type) {
	case *ast.DeclStmt:
		d := s.Decl.(*ast.GenDecl)
		EvalVarDeclBlock(c, d)
	case *ast.AssignStmt:
		for _, l := range s.Lhs {
			switch l.(type) {
			case *ast.Ident:
			default:
				EvalExpr(c, l)
			}
		}
		for _, r := range s.Rhs {
			EvalExpr(c, r)
		}
		for _, l := range s.Lhs {
			ident, isIdent := l.(*ast.Ident)
			if !isIdent {
				continue
			}
			c.scope.MarkInitialized(c.p, ident)
		}
	case *ast.GoStmt:
		for _, arg := range s.Call.Args {
			EvalExpr(c, arg)
		}
		lit, isLit := s.Call.Fun.(*ast.FuncLit)
		if isLit {
			EvalFunc(c.p, c.cfgs, c.cfgs.FuncLit(lit), lit.Type, nil, c.reported)
		} else {
			EvalExpr(c, s.Call)
		}
	default:
		Recurse(c, s)
	}
}

func MergeScopes(block *cfg.Block, inputs []*Scope) *Scope {
	uninitialized := map[VarId]*Decl{}
	for _, depScope := range inputs {
		for vid, branchDecl := range depScope.Uninitialized {
			branchDecl := branchDecl
			utils.UpdateMap(uninitialized, vid, func(v *Decl, exists bool) *Decl {
				if v == nil {
					v = &Decl{
						Name:          branchDecl.Name,
						Changed:       false,
						Uninitialized: map[BranchId]DeclBranch{},
					}
				}
				v.Changed = v.Changed || branchDecl.Changed
				return v
			})
		}
	}
	for _, depScope := range inputs {
		depScope := depScope
		for vid, branchDecl := range depScope.Uninitialized {
			branchDecl := branchDecl
			utils.UpdateMap(uninitialized, vid, func(v *Decl, exists bool) *Decl {
				if v.Changed {
					if branchDecl.Changed {
						// Merge this branch's updated uninitialized set
						utils.MergeMap(v.Uninitialized, branchDecl.Uninitialized)
					} else {
						// No changes in this branch, so identify this branch
						// as one missing initialization compared to other source
						// branches
						v.Uninitialized[depScope.Location] = DeclBranch{
							Comment: depScope.Comment,
						}
					}
				} else {
					// No changes anywhere, use the old uninitialized sets
					utils.MergeMap(v.Uninitialized, branchDecl.Uninitialized)
				}
				return v
			})
		}
	}

	if block == nil {
		return &Scope{
			Location:      0,
			Comment:       "",
			Uninitialized: uninitialized,
		}
	} else {
		return &Scope{
			Location:      BranchId(block.Pos),
			Comment:       BlockComment(block),
			Uninitialized: uninitialized,
		}
	}
}

var extractCommentRegexp = regexp.MustCompile(`\(([^)]*)\)$`)

func BlockComment(b *cfg.Block) string {
	return extractCommentRegexp.FindStringSubmatch(b.String())[1]
}

func CalcDepth(depths map[*cfg.Block]int, b *cfg.Block, depth int) {
	_, hasDepth := depths[b]
	if hasDepth {
		return
	}
	depths[b] = depth
	for _, s := range b.Succs {
		CalcDepth(depths, s, depth+1)
	}
}

func EvalFunc(
	p *analysis.Pass,
	cfgs *ctrlflow.CFGs,
	flow *cfg.CFG,
	spec *ast.FuncType,
	inputs []*Scope,
	reported map[VarId]bool,
) *Scope {
	// Calculate block dependencies and which blocks are actually reachable
	deps := map[*cfg.Block][]*cfg.Block{}
	live := []*cfg.Block{}
	for _, b := range flow.Blocks {
		if !b.Live {
			continue
		}
		utils.Append(&live, b)
		for _, s := range b.Succs {
			if !s.Live {
				continue
			}
			if s == b {
				continue
			}
			deps[s] = append(deps[s], b)
		}
	}

	// Do a dependency ordering of the live blocks
	depths := map[*cfg.Block]int{}
	CalcDepth(depths, flow.Blocks[0], 0)
	done := map[*cfg.Block]bool{}
	orderedBlocks := []*cfg.Block{}
	remaining := live
	for len(remaining) > 0 {
		i := 0
		for {
			if i >= len(remaining) {
				break
			}
			b := remaining[i]
			depsDone := true
			for _, dep := range deps[b] {
				if depths[dep] > depths[b] {
					// Loop (for loop)
					continue
				}
				if !done[dep] {
					depsDone = false
				}
			}
			if !depsDone {
				i += 1
				continue
			}
			utils.Append(&orderedBlocks, b)
			utils.Remove(&remaining, i, 1)
			done[b] = true
		}
	}

	// Outputs is the end scope for any block that the function exits from
	outputs := []*Scope{}
	emptyReturnOutputs := []*Scope{} // blocks with no explicit returns (i.e. "return" not "return 4")

	hasNamedReturns := false
	blockScopes := map[*cfg.Block]*Scope{}
	for bI, b := range orderedBlocks {
		// Aggregate scopes from dependency blocks
		var depScopes []*Scope
		if len(deps[b]) == 0 {
			depScopes = inputs
		} else {
			depScopes = []*Scope{}
			for _, dep := range deps[b] {
				depScope := blockScopes[dep]
				if depScope == nil {
					continue
				}
				utils.Append(&depScopes, blockScopes[dep])
			}
		}
		scope := MergeScopes(b, depScopes)

		// For first block, also add named returns as vars
		if bI == 0 {
			for _, name := range utils.NamedReturns(spec) {
				hasNamedReturns = true
				scope.NewDecl(p, name)
			}
		}

		// Process elements
		c := &Context{
			p:        p,
			cfgs:     cfgs,
			scope:    scope,
			reported: reported,
		}
		for _, e0 := range b.Nodes {
			switch e := e0.(type) {
			case ast.Stmt:
				EvalStmt(c, e)
			case ast.Expr:
				EvalExpr(c, e)
			case *ast.ValueSpec:
				EvalVarDecl(c, e)
			default:
				panic("")
			}
		}

		// Store results
		blockScopes[b] = scope
		if len(b.Succs) == 0 {
			utils.Append(&outputs, scope)
			if len(b.Nodes) > 0 {
				switch l := utils.Last(b.Nodes).(type) {
				case *ast.ReturnStmt:
					if len(l.Results) == 0 {
						utils.Append(&emptyReturnOutputs, scope)
					}
				}
			}
		}
	}

	outScope := MergeScopes(nil, outputs)

	// Check named returns for initialization too
	if hasNamedReturns {
		endContext := &Context{
			p:        p,
			cfgs:     cfgs,
			scope:    MergeScopes(nil, emptyReturnOutputs),
			reported: reported,
		}
		for _, name := range utils.NamedReturns(spec) {
			CheckUseByDecl(endContext, VarId(name.Pos()), name.Name, name.Pos())
		}
	}

	return outScope
}

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "varinit",
		Doc:      "_",
		Requires: []*analysis.Analyzer{ctrlflow.Analyzer},
		Run: func(p *analysis.Pass) (interface{}, error) {
			cfgs := p.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
			reported := map[VarId]bool{}
			for _, file := range p.Files {
				globalScope := &Scope{
					Location:      BranchId(file.Pos()),
					Comment:       "package",
					Uninitialized: map[VarId]*Decl{},
				}
				c := &Context{
					p:        p,
					cfgs:     cfgs,
					scope:    globalScope,
					reported: reported,
				}
				for _, decl := range file.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						flow := cfgs.FuncDecl(d)
						if flow != nil {
							EvalFunc(p, cfgs, flow, d.Type, nil, reported)
						}
					case *ast.GenDecl:
						EvalVarDeclBlock(c, d)
					default:
						panic("UNIMPLEMENTED")
					}
					for v, info := range globalScope.Uninitialized {
						if len(info.Uninitialized) > 0 {
							p.Report(analysis.Diagnostic{
								Pos:     token.Pos(v),
								Message: "This variable was never explicitly initialized",
							})
						}
					}
				}
			}
			return nil, nil
		},
	}
}
