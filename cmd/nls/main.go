package main

import (
	"os"

	"github.com/nolight132/nls/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
