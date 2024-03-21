package optionaok

// check:allfields
type MyStruct struct { // want MyStruct:".*"
	X int `optional:"1"`
}

func consume(x MyStruct) {}

func main() {
	consume(MyStruct{})
}
