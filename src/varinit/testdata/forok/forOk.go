package forok

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	for {
		x = 3
		break
	}
	consume(x)
}
