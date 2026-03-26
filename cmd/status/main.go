package main

import (
	"os"

	"github.com/carlosonunez/status/internal/cli"

	// Register built-in event getters.
	_ "github.com/carlosonunez/status/internal/getter/dummy"

	// Register built-in status setters.
	_ "github.com/carlosonunez/status/internal/setter/dummy"
)

func main() {
	root := cli.NewRootCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
