package switchnodefault

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	switch produce() {
	case 3:
		x = 1
	case 4:
		x = 3
	}
	consume(x) // want "x"
}
