package main

import (
	"os"

	"mort/internal/xtbmlcli"
)

func main() {
	code := xtbmlcli.Run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}
