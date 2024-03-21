package fornobreak

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	for produce() != 7 {
		x = 3
	}
	consume(x) // want "x"
}
