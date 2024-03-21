package primcastok

type T int

func consume(t T) {}

func main() {
	consume(T(4))
}
