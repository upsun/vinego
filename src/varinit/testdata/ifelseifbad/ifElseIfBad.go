package ifelseifbad

func produce() int  { return 3 }
func consume(x int) {}

func main() {
	var x int
	if produce() == 3 {
		produce()
	} else {
		x = 4
	}
	consume(x) // want "x"
}
