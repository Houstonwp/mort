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
	var doc struct {
		Tables []struct {
			Meta struct {
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
			} `xml:"MetaData"`
		} `xml:"Table"`
	}

	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode table metas: %w", err)
	}

	metas := make([]TableMeta, 0, len(doc.Tables))
	for _, table := range doc.Tables {
		meta := table.Meta
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

		metas = append(metas, TableMeta{
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
		})
	}

	return metas, nil
}
