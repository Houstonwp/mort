package xtbmlcli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"mort/internal/xtbml"
)

// Run executes the converter CLI with the provided arguments.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("xtbmlconvert", flag.ContinueOnError)
	fs.SetOutput(stderr)

	src := fs.String("src", "xml", "directory containing XTbML XML files")
	dst := fs.String("dst", "json", "directory for JSON output")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	var converted int
	err := xtbml.ConvertDirectoryWithObserver(*src, *dst, func(srcPath, dstPath string) {
		fmt.Fprintf(stdout, "Converted %s -> %s\n", filepath.Base(srcPath), dstPath)
		converted++
	})
	if err != nil {
		fmt.Fprintf(stderr, "conversion failed: %v\n", err)
		return 1
	}

	if converted == 0 {
		fmt.Fprintln(stdout, "No XML files converted.")
	}
	return 0
}
