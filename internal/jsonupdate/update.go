// Package jsonupdate keeps changed mortality XML and generated JSON in sync.
package jsonupdate

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"mort/internal/xtbml"
)

type update struct {
	path   string
	data   []byte
	remove bool
}

// Update regenerates tables changed since base, including staged and untracked
// files. In check mode it reports stale JSON without writing any repository files.
// root must be the repository root. The changelog cursor is not a table.
func Update(root, base string, check bool, stdout io.Writer) error {
	sources, err := changedSources(root, base)
	if err != nil {
		return err
	}
	var updates []update
	for _, source := range sources {
		change, err := prepareUpdate(root, source)
		if err != nil {
			return err
		}
		if change != nil {
			updates = append(updates, *change)
		}
	}
	if check && len(updates) > 0 {
		for _, change := range updates {
			fmt.Fprintf(stdout, "Stale generated file: %s\n", change.path)
		}
		return fmt.Errorf("%d generated files need updating; run go run ./cmd/updatejson -base %q", len(updates), base)
	}
	for _, change := range updates {
		if err := applyUpdate(root, change); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Updated %s\n", change.path)
	}
	fmt.Fprintf(stdout, "Checked %d changed table(s).\n", len(sources))
	return nil
}

func changedSources(root, base string) ([]string, error) {
	ref, err := git(root, "rev-parse", "--verify", "--end-of-options", base+"^{commit}")
	if err != nil {
		return nil, err
	}
	tracked, err := git(root, "diff", "--name-only", "-z", "--no-renames", strings.TrimSpace(string(ref)), "--", "xml", "json")
	if err != nil {
		return nil, err
	}
	untracked, err := git(root, "ls-files", "--others", "--exclude-standard", "-z", "--", "xml", "json")
	if err != nil {
		return nil, err
	}
	unique := make(map[string]bool)
	for _, name := range strings.Split(string(tracked)+string(untracked), "\x00") {
		switch {
		case strings.HasPrefix(name, "xml/") && strings.HasSuffix(name, ".xml"):
			unique[name] = true
		case strings.HasPrefix(name, "json/") && strings.HasSuffix(name, ".json") && name != "json/changelog_state.json":
			unique["xml/"+strings.TrimSuffix(strings.TrimPrefix(name, "json/"), ".json")+".xml"] = true
		}
	}
	sources := make([]string, 0, len(unique))
	for source := range unique {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	return sources, nil
}

func prepareUpdate(root, source string) (*update, error) {
	destination := "json/" + strings.TrimSuffix(strings.TrimPrefix(source, "xml/"), ".xml") + ".json"
	current, err := readRegularFile(filepath.Join(root, destination))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	outputExists := err == nil
	raw, err := readRegularFile(filepath.Join(root, source))
	if os.IsNotExist(err) {
		if !outputExists {
			return nil, nil
		}
		return &update{path: destination, remove: true}, nil
	}
	if err != nil {
		return nil, err
	}
	converted, err := xtbml.ConvertXTbml(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("convert %s: %w", source, err)
	}
	if outputExists && bytes.Equal(current, converted) {
		return nil, nil
	}
	return &update{path: destination, data: converted}, nil
}

func readRegularFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("expected a regular file: %s", path)
	}
	return os.ReadFile(path)
}

func applyUpdate(root string, change update) error {
	path := filepath.Join(root, change.path)
	if change.remove {
		return os.Remove(path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, change.data, 0o644)
}

func git(root string, args ...string) ([]byte, error) {
	output, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, output)
	}
	return output, nil
}
