package forinitok

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	for x = 3; produce() == 12; {
		break
	}
	consume(x)
}
