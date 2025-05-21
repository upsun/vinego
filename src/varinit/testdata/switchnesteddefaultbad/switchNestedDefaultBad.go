package switchnesteddefaultbad

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	switch produce() {
	case 3:
		//
		x = 2
	case 4:
		switch produce() {
		case 5:
			x = 7
		}
	default:
		x = 4
	}
	consume(x) // want "x"
}
