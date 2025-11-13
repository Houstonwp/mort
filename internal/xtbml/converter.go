package xtbml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// ConvertXTbml reads an XTbML XML payload and returns normalized JSON bytes.
func ConvertXTbml(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	version, err := InferVersion(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	classification, err := ParseContentClassification(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	tableMetas, err := ParseTableMetas(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	points, err := ParseRates(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	payload := struct {
		Identifier     string                 `json:"identifier"`
		Version        string                 `json:"version"`
		Classification *ClassificationPayload `json:"classification"`
		Tables         []TablePayload         `json:"tables"`
	}{
		Identifier:     NormalizeIdentifier(classification.TableName),
		Version:        version,
		Classification: toClassificationPayload(classification),
	}

	tableMap := make(map[int][]RateEntryPayload)
	for _, point := range points {
		entry := RateEntryPayload{
			Age:      point.Age,
			Duration: point.Duration,
			Rate:     point.Rate,
		}
		tableMap[point.Table] = append(tableMap[point.Table], entry)
	}

	if len(tableMap) > 0 {
		indexes := make([]int, 0, len(tableMap))
		for idx := range tableMap {
			indexes = append(indexes, idx)
		}
		sort.Ints(indexes)

		payload.Tables = make([]TablePayload, len(indexes))
		for i, idx := range indexes {
			payload.Tables[i] = TablePayload{
				Index:    idx,
				Metadata: metaForIndex(idx, tableMetas),
				Rates:    tableMap[idx],
			}
		}
	}

	if len(payload.Tables) == 0 && len(tableMetas) > 0 {
		payload.Tables = make([]TablePayload, len(tableMetas))
		for i := range tableMetas {
			payload.Tables[i] = TablePayload{
				Index:    i,
				Metadata: metaPayload(tableMetas[i]),
			}
		}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return buf.Bytes(), nil
}

type ClassificationPayload struct {
	TableIdentity    string                 `json:"tableIdentity"`
	ProviderDomain   string                 `json:"providerDomain"`
	ProviderName     string                 `json:"providerName"`
	TableReference   string                 `json:"tableReference"`
	ContentType      ClassifiedValuePayload `json:"contentType"`
	TableName        string                 `json:"tableName"`
	TableDescription string                 `json:"tableDescription"`
	Comments         string                 `json:"comments"`
	Keywords         []string               `json:"keywords"`
}

type TablePayload struct {
	Index    int                `json:"index"`
	Metadata *TableMetaPayload  `json:"metadata,omitempty"`
	Rates    []RateEntryPayload `json:"rates,omitempty"`
}

type TableMetaPayload struct {
	ScalingFactor    string                  `json:"scalingFactor"`
	DataType         ClassifiedValuePayload  `json:"dataType"`
	Nation           ClassifiedValuePayload  `json:"nation"`
	TableDescription string                  `json:"tableDescription"`
	Axes             []AxisDefinitionPayload `json:"axes"`
}

type AxisDefinitionPayload struct {
	ID        string                 `json:"id"`
	ScaleType ClassifiedValuePayload `json:"scaleType"`
	AxisName  string                 `json:"axisName"`
	MinValue  string                 `json:"minValue"`
	MaxValue  string                 `json:"maxValue"`
	Increment string                 `json:"increment"`
}

type ClassifiedValuePayload struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type RateEntryPayload struct {
	Age      int      `json:"age"`
	Duration *int     `json:"duration,omitempty"`
	Rate     *float64 `json:"rate"`
}

func toClassificationPayload(class *ContentClassification) *ClassificationPayload {
	if class == nil {
		return nil
	}
	return &ClassificationPayload{
		TableIdentity:    class.TableIdentity,
		ProviderDomain:   class.ProviderDomain,
		ProviderName:     class.ProviderName,
		TableReference:   class.TableReference,
		ContentType:      ClassifiedValuePayload{Code: class.ContentType.Code, Label: class.ContentType.Label},
		TableName:        class.TableName,
		TableDescription: class.TableDescription,
		Comments:         class.Comments,
		Keywords:         class.Keywords,
	}
}

func metaForIndex(idx int, metas []TableMeta) *TableMetaPayload {
	if idx < 0 || idx >= len(metas) {
		return nil
	}
	return metaPayload(metas[idx])
}

func metaPayload(meta TableMeta) *TableMetaPayload {
	return &TableMetaPayload{
		ScalingFactor:    meta.ScalingFactor,
		DataType:         ClassifiedValuePayload{Code: meta.DataType.Code, Label: meta.DataType.Label},
		Nation:           ClassifiedValuePayload{Code: meta.Nation.Code, Label: meta.Nation.Label},
		TableDescription: meta.TableDescription,
		Axes:             axisPayloads(meta.Axes),
	}
}

func axisPayloads(axes []AxisDefinition) []AxisDefinitionPayload {
	out := make([]AxisDefinitionPayload, len(axes))
	for i, axis := range axes {
		out[i] = AxisDefinitionPayload{
			ID: axis.ID,
			ScaleType: ClassifiedValuePayload{
				Code:  axis.ScaleType.Code,
				Label: axis.ScaleType.Label,
			},
			AxisName:  axis.AxisName,
			MinValue:  axis.MinValue,
			MaxValue:  axis.MaxValue,
			Increment: axis.Increment,
		}
	}
	return out
}
