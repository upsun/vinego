package namedreturnbad

func MyFunc() (x int, z string) { // want "z"
	x = 4
	return
}
