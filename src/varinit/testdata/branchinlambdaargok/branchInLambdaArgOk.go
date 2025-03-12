package lambdaargok

func Func(func()) {
}

func produce() int {
	return 99
}

func consume(int) {
}

func main() {
	Func(func() {
		var x int
		if produce() > 0 {
			x = 7
		} else {
			x = 4
		}

		if produce() > 10 {
			consume(x)
		}
	})
}
