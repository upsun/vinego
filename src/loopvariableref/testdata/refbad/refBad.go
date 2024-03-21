package refbad

import "log"

func main() {
	x := []int{4, 5, 6}
	for k, v := range x {
		log.Printf("%p %p", &k, &v) // want "Risky.*variable k" "Risky.*variable v"
	}
}
