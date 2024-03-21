package forbad

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	for {
		break
	}
	consume(x) // want "x"
}
