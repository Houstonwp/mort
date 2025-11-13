package xtbml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTableMetas(t *testing.T) {
	path := filepath.Join("testdata", "table_meta.xml")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	metas, err := ParseTableMetas(f)
	if err != nil {
		t.Fatalf("ParseTableMetas() error = %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("expected 2 table metas, got %d", len(metas))
	}
	if metas[0].TableDescription != "Ultimate rates." || len(metas[0].Axes) != 1 {
		t.Fatalf("first meta mismatch: %#v", metas[0])
	}
	if metas[1].TableDescription != "Select rates." || len(metas[1].Axes) != 2 {
		t.Fatalf("second meta mismatch: %#v", metas[1])
	}
	if metas[1].Axes[1].ID != "Duration" || metas[1].Axes[1].ScaleType.Label != "Duration" {
		t.Fatalf("duration axis mismatch: %#v", metas[1].Axes[1])
	}
}
