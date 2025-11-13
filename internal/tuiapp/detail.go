package tuiapp

import (
	"encoding/json"
	"fmt"
	"os"
)

// TableDetail mirrors the JSON schema for UI consumption.
type TableDetail struct {
	Identifier     string                `json:"identifier"`
	Version        string                `json:"version"`
	Classification ClassificationPayload `json:"classification"`
	Tables         []TablePayload        `json:"tables"`
}

type ClassificationPayload struct {
	TableIdentity    string       `json:"tableIdentity"`
	ProviderDomain   string       `json:"providerDomain"`
	ProviderName     string       `json:"providerName"`
	TableReference   string       `json:"tableReference"`
	ContentType      ValuePayload `json:"contentType"`
	TableName        string       `json:"tableName"`
	TableDescription string       `json:"tableDescription"`
	Comments         string       `json:"comments"`
	Keywords         []string     `json:"keywords"`
}

type TablePayload struct {
	Index    int                `json:"index"`
	Metadata *TableMetaPayload  `json:"metadata"`
	Rates    []RateEntryPayload `json:"rates"`
}

type TableMetaPayload struct {
	ScalingFactor    string        `json:"scalingFactor"`
	DataType         ValuePayload  `json:"dataType"`
	Nation           ValuePayload  `json:"nation"`
	TableDescription string        `json:"tableDescription"`
	Axes             []AxisPayload `json:"axes"`
}

type AxisPayload struct {
	ID        string       `json:"id"`
	ScaleType ValuePayload `json:"scaleType"`
	AxisName  string       `json:"axisName"`
	MinValue  string       `json:"minValue"`
	MaxValue  string       `json:"maxValue"`
	Increment string       `json:"increment"`
}

type ValuePayload struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type RateEntryPayload struct {
	Age      int      `json:"age"`
	Duration *int     `json:"duration"`
	Rate     *float64 `json:"rate"`
}

// LoadTableDetail reads a single JSON file into a TableDetail structure.
func LoadTableDetail(path string) (*TableDetail, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var detail TableDetail
	if err := json.Unmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &detail, nil
}
