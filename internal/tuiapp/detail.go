package tuiapp

import (
	"encoding/json"
	"fmt"
	"os"

	"mort/internal/xtbml"
)

// TableDetail mirrors the converter's JSON payload for UI consumption.
type TableDetail = xtbml.ConvertedTable

// LoadTableDetail reads a single JSON file into a TableDetail structure.
func LoadTableDetail(path string) (*TableDetail, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var detail xtbml.ConvertedTable
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &detail, nil
}
