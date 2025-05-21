package switchdefaultbad

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	switch produce() {
	case 3:
		x = 2
	case 4:
		x = 7
	default:
	}
	consume(x) // want "x"
}
