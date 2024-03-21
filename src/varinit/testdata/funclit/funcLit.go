package funclit

func produce() int  { return 7 }
func consume(x int) {}

func main() {
	var x int
	func() {
		x = 7
	}()
	consume(x)
}
