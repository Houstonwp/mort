// updatejson regenerates or checks JSON for XML/JSON changes relative to a commit.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mort/internal/jsonupdate"
)

func main() {
	base := flag.String("base", "HEAD", "base commit for changed XML and JSON files")
	check := flag.Bool("check", false, "fail on stale JSON without changing files")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		os.Exit(2)
	}
	root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		err = jsonupdate.Update(strings.TrimSpace(string(root)), *base, *check, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
