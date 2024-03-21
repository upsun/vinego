package bad

import "fmt"

func bad() error {
	return fmt.Errorf("xox")
}

func main() {
	var err error
	func() {
		err = bad() // want "Assigning to captured err variable"
	}()
	_ = err
}
