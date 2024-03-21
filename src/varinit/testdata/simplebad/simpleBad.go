package simplebad

func consume(i int) {}

func main() {
	var x int
	consume(x) // want "x"
}
