package xtbml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertXTbml_Golden(t *testing.T) {
	xmlPath := filepath.Join("testdata", "table_small.xml")
	jsonPath := filepath.Join("testdata", "json", "table_small.json")

	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		t.Fatalf("open xml fixture: %v", err)
	}
	defer xmlFile.Close()

	gotBytes, err := ConvertXTbml(xmlFile)
	if err != nil {
		t.Fatalf("ConvertXTbml() error = %v", err)
	}

	wantBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read golden json: %v", err)
	}

	var got, want any
	if err := json.Unmarshal(gotBytes, &got); err != nil {
		t.Fatalf("json.Unmarshal got: %v\n%s", err, string(gotBytes))
	}
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatalf("json.Unmarshal want: %v", err)
	}
	if !equalJSON(got, want) {
		t.Fatalf("ConvertXTbml() mismatch\n got: %s\nwant: %s", string(gotBytes), string(wantBytes))
	}
}

func equalJSON(a, b any) bool {
	return jsonDeepEqual(a, b)
}

func jsonDeepEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !jsonDeepEqual(v, bv[k]) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonDeepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return av == b
	}
}
