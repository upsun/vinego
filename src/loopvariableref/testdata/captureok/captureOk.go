package captureok

import "log"

func main() {
	x := 7
	func() {
		log.Printf("%d", x)
	}()
}
