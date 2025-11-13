package xtbml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInferVersion(t *testing.T) {
	t.Run("extracts version attribute", func(t *testing.T) {
		path := filepath.Join("testdata", "basic_version.xml")
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer f.Close()

		got, err := InferVersion(f)
		if err != nil {
			t.Fatalf("InferVersion() unexpected error: %v", err)
		}
		if got != "1.3" {
			t.Fatalf("InferVersion() = %q, want %q", got, "1.3")
		}
	})

	t.Run("fails when attribute missing", func(t *testing.T) {
		xml := []byte(`<XTbML></XTbML>`)
		got, err := InferVersion(bytes.NewReader(xml))
		if err != nil {
			t.Fatalf("InferVersion() unexpected error: %v", err)
		}
		if got != "unknown" {
			t.Fatalf("InferVersion() = %q, want unknown fallback", got)
		}
	})
}
