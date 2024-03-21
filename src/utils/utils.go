package utils

import (
	"go/ast"
)

func P[V any](v V) *V {
	return &v
}

func Append[E any](d *[]E, v ...E) {
	*d = append(*d, v...)
}

func Remove[E any](d *[]E, start int, count int) {
	*d = append((*d)[:start], (*d)[start+count:]...)
}

func PopMap[K comparable, V any](m map[K]V, key K) V {
	out := m[key]
	delete(m, key)
	return out
}

func UpdateMap[K comparable, V any](
	m map[K]V,
	key K,
	updater func(v V, exists bool) V,
) {
	old, exists := m[key]
	m[key] = updater(old, exists)
}

func MergeMap[K comparable, V any](
	dest map[K]V,
	source map[K]V,
) {
	for k, v := range source {
		dest[k] = v
	}
}

func Last[V any](v []V) V {
	return v[len(v)-1]
}

// ArrAntiSuffix returns all but count last elements
func ArrAntiSuffix[V any](v []V, count int) []V {
	return v[:len(v)-count]
}

type walkPair[S any] struct {
	Node       ast.Node
	StoredData S
}

func Walk[S any](n ast.Node, preVisitor func(n ast.Node) (S, bool), postVisitor func(stored S, n ast.Node)) {
	stack := []walkPair[S]{}
	ast.Inspect(n, func(n ast.Node) bool {
		if n != nil {
			// Enter node
			s, out := preVisitor(n)
			if out {
				Append(&stack, walkPair[S]{
					Node:       n,
					StoredData: s,
				})
			}
			return out
		} else {
			// Exit node
			lastI := len(stack) - 1
			top := stack[lastI]
			postVisitor(top.StoredData, top.Node)
			stack = stack[:lastI]
			return true
		}
	})
}

func WalkWithCrumbs(n ast.Node, visitor func(n ast.Node, crumbs []ast.Node) bool) {
	stack := []ast.Node{}
	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		out := visitor(n, stack)
		Append(&stack, n)
		return out
	})
}

/*
DEBUG TOOLS

Uncomment and use to print debug stuff, then read it with
./build.py exec -- bash -c '/bin/golangci-lint run --timeout 10m -v --new; cat falselog.txt'

(because golangci-lint swallows all stdout/stderr writes)

var fl *os.File

func init() {
	var err error
	fl, err = os.Create("falselog.txt")
	if err != nil {
		panic(err)
	}
}

func FalseLog(pattern string, args ...any) {
	_, _ = fl.WriteString(fmt.Sprintf(pattern, args...) + "\n")
	_ = fl.Sync()
}
*/

func NamedReturns(spec *ast.FuncType) []*ast.Ident {
	out := []*ast.Ident{}
	if spec.Results != nil {
		for _, inject := range spec.Results.List {
			out = append(out, inject.Names...)
		}
	}
	return out
}
