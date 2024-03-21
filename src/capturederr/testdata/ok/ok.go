package ok

import "fmt"

func bad() error {
	return fmt.Errorf("xox")
}

func main() {
	var err error
	func() {
		err := bad()
		_ = err
	}()
	_ = err
}
