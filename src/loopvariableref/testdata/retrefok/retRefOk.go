package retrefok

func MyFunc() *int {
	x := []int{4, 6, 7}
	for _, v := range x {
		return &v
	}
}
