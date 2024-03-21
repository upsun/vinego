package argok

import "fmt"

func bad() error {
	return fmt.Errorf("xox")
}

func handle(err error) {
	err = bad()
}
