package main

import (
	"os"
	"transaction-indexer/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
