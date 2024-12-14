package globalconst

const (
	Foo = iota
	Bar
	Baz
)

func consume(x int) {}

func main() {
	consume(Foo)
	consume(Bar)
	consume(Baz)
}
