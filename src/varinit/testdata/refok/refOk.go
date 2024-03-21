package refok

func consume(i int) {}

func outvarinit(i *int) {}

func main() {
	var x int
	outvarinit(&x)
	consume(x)
}
