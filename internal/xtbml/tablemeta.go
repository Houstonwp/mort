package xtbml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// TableMeta represents metadata for a specific <Table>.
type TableMeta struct {
	ScalingFactor    string
	DataType         ClassifiedValue
	Nation           ClassifiedValue
	TableDescription string
	Axes             []AxisDefinition
}

// AxisDefinition describes an <AxisDef> entry.
type AxisDefinition struct {
	ID        string
	ScaleType ClassifiedValue
	AxisName  string
	MinValue  string
	MaxValue  string
	Increment string
}

// ParseTableMetas extracts the metadata blocks for every <Table>.
func ParseTableMetas(r io.Reader) ([]TableMeta, error) {
	dec := xml.NewDecoder(r)
	var (
		metas      []TableMeta
		tableIndex = -1
	)

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode table metas: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch strings.ToLower(start.Name.Local) {
		case "table":
			tableIndex++
			metas = append(metas, TableMeta{})
		case "metadata":
			if tableIndex < 0 {
				continue
			}
			meta, err := decodeMetaElement(dec, start)
			if err != nil {
				return nil, err
			}
			metas[tableIndex] = meta
		}
	}

	return metas, nil
}

func decodeMetaElement(dec *xml.Decoder, start xml.StartElement) (TableMeta, error) {
	var meta struct {
		ScalingFactor string `xml:"ScalingFactor"`
		DataType      struct {
			Code  string `xml:"tc,attr"`
			Label string `xml:",chardata"`
		} `xml:"DataType"`
		Nation struct {
			Code  string `xml:"tc,attr"`
			Label string `xml:",chardata"`
		} `xml:"Nation"`
		TableDescription string `xml:"TableDescription"`
		Axes             []struct {
			ID        string `xml:"id,attr"`
			ScaleType struct {
				Code  string `xml:"tc,attr"`
				Label string `xml:",chardata"`
			} `xml:"ScaleType"`
			AxisName  string `xml:"AxisName"`
			MinValue  string `xml:"MinScaleValue"`
			MaxValue  string `xml:"MaxScaleValue"`
			Increment string `xml:"Increment"`
		} `xml:"AxisDef"`
	}
	if err := dec.DecodeElement(&meta, &start); err != nil {
		return TableMeta{}, fmt.Errorf("decode table metas: %w", err)
	}

	axes := make([]AxisDefinition, 0, len(meta.Axes))
	for _, axis := range meta.Axes {
		axes = append(axes, AxisDefinition{
			ID: axis.ID,
			ScaleType: ClassifiedValue{
				Code:  strings.TrimSpace(axis.ScaleType.Code),
				Label: strings.TrimSpace(axis.ScaleType.Label),
			},
			AxisName:  strings.TrimSpace(axis.AxisName),
			MinValue:  strings.TrimSpace(axis.MinValue),
			MaxValue:  strings.TrimSpace(axis.MaxValue),
			Increment: strings.TrimSpace(axis.Increment),
		})
	}

	return TableMeta{
		ScalingFactor: strings.TrimSpace(meta.ScalingFactor),
		DataType: ClassifiedValue{
			Code:  strings.TrimSpace(meta.DataType.Code),
			Label: strings.TrimSpace(meta.DataType.Label),
		},
		Nation: ClassifiedValue{
			Code:  strings.TrimSpace(meta.Nation.Code),
			Label: strings.TrimSpace(meta.Nation.Label),
		},
		TableDescription: strings.TrimSpace(meta.TableDescription),
		Axes:             axes,
	}, nil
}
