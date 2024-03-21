package forlabelbreak

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
Outer:
	for {
		for {
			x = 3
			break Outer
		}
	}
	consume(x)
}
