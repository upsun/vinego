package primcastbad

type T int

func consume(t T) {}

func main() {
	consume(4) // want "Implicit"
}
