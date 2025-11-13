package xtbmlcli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	xmlBytes, err := os.ReadFile(filepath.Join("..", "xtbml", "testdata", "table_small.xml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "table_small.xml"), xmlBytes, 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--src", src, "--dst", dst}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("table_small.xml")) {
		t.Fatalf("stdout missing file name: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr not empty: %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dst, "table_small.json")); err != nil {
		t.Fatalf("expected output json: %v", err)
	}
}

func TestRunFailure(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := os.WriteFile(filepath.Join(src, "broken.xml"), []byte(`<XTbML>`), 0o644); err != nil {
		t.Fatalf("write broken xml: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--src", src, "--dst", dst}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected failure exit code, got 0")
	}
	if stderr.Len() == 0 {
		t.Fatalf("expected stderr output")
	}
}
