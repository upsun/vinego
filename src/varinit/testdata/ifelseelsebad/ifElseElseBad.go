package ifelseelsebad

func produce() int  { return 3 }
func consume(x int) {}

func main() {
	var x int
	if produce() == 3 {
		x = 7
	} else {
		produce()
	}
	consume(x) // want "x"
}
