package main

import (
	"errors"
	"os"

	"github.com/nolight132/nls/internal/cli"
	"github.com/nolight132/nls/internal/output"
)

func main() {
	if err := cli.Root().Execute(); err != nil {
		if !errors.Is(err, cli.ErrReported) {
			output.WriteError(err, output.StderrIsTTY())
		}
		os.Exit(1)
	}
}
