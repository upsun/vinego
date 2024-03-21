package namelssok

// check:allfields
type MyStruct struct { // want MyStruct:".*"
	X int
}

func consume(x []*MyStruct) {}

func main() {
	consume([]*MyStruct{
		{
			X: 4,
		},
	})
}
