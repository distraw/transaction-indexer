package main

import (
	"os"

	"github.com/distraw/transaction-indexer/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
