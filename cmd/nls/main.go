package main

import (
	"os"

	"github.com/nolight132/nls/internal/cli"
)

func main() {
	if err := cli.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
