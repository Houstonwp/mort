package xtbml

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ConvertDirectory walks srcDir for *.xml files and writes JSON outputs to dstDir.
func ConvertDirectory(srcDir, dstDir string) error {
	return ConvertDirectoryWithObserver(srcDir, dstDir, nil)
}

// ConvertDirectoryWithObserver mirrors ConvertDirectory and reports each conversion via observer.
func ConvertDirectoryWithObserver(srcDir, dstDir string, observer func(src, dst string)) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read src dir: %w", err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("ensure dst dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".xml") {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		dstName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) + ".json"
		dstPath := filepath.Join(dstDir, dstName)

		if err := ConvertFile(srcPath, dstPath); err != nil {
			return err
		}
		if observer != nil {
			observer(srcPath, dstPath)
		}
	}

	return nil
}

// ConvertFile converts a single XML file at srcPath into JSON at dstPath.
func ConvertFile(srcPath, dstPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}
	out, err := ConvertXTbml(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("convert %s: %w", srcPath, err)
	}
	if err := os.WriteFile(dstPath, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dstPath, err)
	}
	return nil
}
