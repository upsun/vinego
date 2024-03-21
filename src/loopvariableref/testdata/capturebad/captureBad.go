package capturebad

import "log"

func main() {
	x := []int{4, 6, 7}
	for k, v := range x {
		func() {
			log.Printf("%d %d", k, v) // want "Risky.*variable k" "Risky.*variable v"
		}()
	}
}
