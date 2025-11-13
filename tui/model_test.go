package tui

import (
	"path/filepath"
	"testing"
)

func TestLoadSummariesCmd(t *testing.T) {
	dir := filepath.Join("..", "internal", "tuiapp", "testdata", "json")
	cmd := loadSummariesCmd(dir, nil, 0)
	msg := cmd()
	first, ok := msg.(summariesChunkMsg)
	if !ok {
		t.Fatalf("expected summariesChunkMsg, got %T", msg)
	}
	if first.err != nil {
		t.Fatalf("chunk error: %v", first.err)
	}
	if len(first.chunk) == 0 {
		t.Fatalf("expected non-empty chunk")
	}
	cmd = loadSummariesCmd(dir, first.files, first.next)
	secondMsg := cmd()
	second, ok := secondMsg.(summariesChunkMsg)
	if !ok {
		t.Fatalf("expected summariesChunkMsg, got %T", secondMsg)
	}
	if second.err != nil {
		t.Fatalf("chunk error: %v", second.err)
	}
	if !second.done {
		t.Fatalf("expected done after second chunk")
	}
}

func TestLoadDetailCmd(t *testing.T) {
	path := filepath.Join("..", "internal", "tuiapp", "testdata", "json", "table_alpha.json")
	cmd := loadDetailCmd(path)
	msg := cmd()
	loaded, ok := msg.(detailLoadedMsg)
	if !ok {
		t.Fatalf("expected detailLoadedMsg, got %T", msg)
	}
	if loaded.err != nil {
		t.Fatalf("detail load error: %v", loaded.err)
	}
	if loaded.detail == nil || loaded.detail.Classification.TableName != "Alpha Table" {
		t.Fatalf("unexpected detail: %#v", loaded.detail)
	}
}
