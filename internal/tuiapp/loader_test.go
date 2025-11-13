package tuiapp

import (
	"path/filepath"
	"testing"
)

func TestLoadTableSummaries(t *testing.T) {
	dir := filepath.Join("testdata", "json")
	summaries, err := LoadTableSummaries(dir)
	if err != nil {
		t.Fatalf("LoadTableSummaries() error = %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].TableIdentity != "alpha" || summaries[0].Name != "Alpha Table" {
		t.Fatalf("first summary mismatch: %#v", summaries[0])
	}
	if summaries[1].TableIdentity != "beta" {
		t.Fatalf("expected beta second: %#v", summaries[1])
	}
}

func TestFilterSummaries(t *testing.T) {
	items := []TableSummary{
		{Name: "2015 VBT Table", Identifier: "table_alpha", TableIdentity: "alpha", Provider: "Provider A", Keywords: []string{"2015", "vbt"}},
		{Name: "Beta Table", Identifier: "table_beta", TableIdentity: "beta"},
		{Name: "2015 Experience Study", Identifier: "table_gamma", TableIdentity: "gamma", Provider: "Provider C"},
	}

	filtered := FilterSummaries(items, "beta")
	if len(filtered) == 0 || filtered[0].Identifier != "table_beta" {
		t.Fatalf("FilterSummaries() provider/name match failed: %#v", filtered)
	}

	filtered = FilterSummaries(items, "2015 vbt")
	if len(filtered) != 1 || filtered[0].Identifier != "table_alpha" {
		t.Fatalf("FilterSummaries() multi-token match failed: %#v", filtered)
	}

	filtered = FilterSummaries(items, "")
	if len(filtered) != len(items) {
		t.Fatalf("empty query should return all")
	}
}
