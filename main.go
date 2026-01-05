package main

import (
	"fmt"
	"os"

	"github.com/bobbyunknown/Oh-my-builder/cmd/omb"
)

func main() {
	if err := omb.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
