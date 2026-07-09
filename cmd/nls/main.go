package main

import (
	"os"

	"github.com/nolight132/nls/internal/cli"
	"github.com/nolight132/nls/internal/output"
)

func main() {
	if err := cli.Root().Execute(); err != nil {
		output.WriteError(err, output.StderrIsTTY())
		os.Exit(1)
	}
}
