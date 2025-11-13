package xtbml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertDirectory(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	xmlBytes, err := os.ReadFile(filepath.Join("testdata", "table_small.xml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "table_small.xml"), xmlBytes, 0o644); err != nil {
		t.Fatalf("write src xml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "notes.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatalf("write extra file: %v", err)
	}

	if err := ConvertDirectory(src, dst); err != nil {
		t.Fatalf("ConvertDirectory() error = %v", err)
	}

	gotPath := filepath.Join(dst, "table_small.json")
	if _, err := os.Stat(gotPath); err != nil {
		t.Fatalf("expected json file, got stat error: %v", err)
	}
	gotBytes, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	wantBytes, err := ConvertXTbml(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("ConvertXTbml() golden error: %v", err)
	}
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("json mismatch\n got: %s\nwant: %s", gotBytes, wantBytes)
	}
}
