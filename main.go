package main

import (
	"fmt"
	"os"

	"github.com/maxbeizer/gh-fleet/cmd"
)

const version = "0.1.0"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
