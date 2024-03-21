package switchok

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	switch produce() {
	case 4:
		x = 7
	default:
		x = 12
	}
	consume(x)
}
