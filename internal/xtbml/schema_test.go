package xtbml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestSampleJSONMatchesSchema(t *testing.T) {
	schemaPath := filepath.Join("..", "..", "schemas", "xtbml.schema.json")
	jsonPath := filepath.Join("testdata", "json", "table_small.json")

	absSchema, err := filepath.Abs(schemaPath)
	if err != nil {
		t.Fatalf("abs schema: %v", err)
	}
	schemaURL := "file://" + filepath.ToSlash(absSchema)

	schema, err := jsonschema.Compile(schemaURL)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}

	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if err := schema.Validate(payload); err != nil {
		t.Fatalf("schema validation failed: %v", err)
	}
}
