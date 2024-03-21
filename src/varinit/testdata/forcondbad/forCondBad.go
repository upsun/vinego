package forcondbad

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	for produce() == 3 {
		x = 3
		break
	}
	consume(x) // want "x"
}
