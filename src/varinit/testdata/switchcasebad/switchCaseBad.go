package switchcasebad

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	switch produce() {
	case 4:
	default:
		x = 3
	}
	consume(x) // want "x"
}
