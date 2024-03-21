package nocaptureok

import "fmt"

func bad() error {
	return fmt.Errorf("xox")
}

func main() {
	var err error
	err = bad()
	_ = err
}
