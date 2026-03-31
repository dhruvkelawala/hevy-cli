package main

import (
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/cmd"
	"github.com/fatih/color"
)

func main() {
	if err := cmd.Execute(); err != nil {
		color.New(color.FgRed).Fprintln(os.Stderr, fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}
