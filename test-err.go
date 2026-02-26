package main

import (
"fmt"
"os"
)

func main() {
	err := fmt.Errorf("this is a test error\n\n  Please run something else.")
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
