package tuiapp

import (
	"path/filepath"
	"testing"
)

func TestLoadTableDetail(t *testing.T) {
	path := filepath.Join("testdata", "json", "table_alpha.json")
	detail, err := LoadTableDetail(path)
	if err != nil {
		t.Fatalf("LoadTableDetail() error = %v", err)
	}
	if detail.Classification.TableName != "Alpha Table" {
		t.Fatalf("unexpected name: %#v", detail.Classification)
	}
	if len(detail.Tables) != 1 || len(detail.Tables[0].Rates) != 1 {
		t.Fatalf("expected rates in detail: %#v", detail.Tables)
	}
}
