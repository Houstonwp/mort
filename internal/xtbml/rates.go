package xtbml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RatePoint captures a single rate entry, optionally scoped to a duration.
type RatePoint struct {
	Table    int
	Age      int
	Duration *int
	Rate     *float64
}

// ParseRates reads XTbML <Values> blocks, supporting both single-axis and nested axes.
func ParseRates(r io.Reader) ([]RatePoint, error) {
	dec := xml.NewDecoder(r)

	var (
		points        []RatePoint
		tableIndex    = -1
		inValues      bool
		axisDepth     int
		hasCurrentAge bool
		currentAge    int
	)

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode rates: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			switch name {
			case "table":
				tableIndex++
			case "values":
				if tableIndex >= 0 {
					inValues = true
					axisDepth = 0
					hasCurrentAge = false
				}
			case "axis":
				if !inValues {
					continue
				}
				axisDepth++
				if axisDepth == 1 {
					if attr, ok := attrInt(t.Attr, "t"); ok {
						currentAge = attr
						hasCurrentAge = true
					} else {
						hasCurrentAge = false
					}
				}
			case "y":
				if !inValues {
					continue
				}
				valueAttr, hasAttr := attrInt(t.Attr, "t")
				var text string
				if err := dec.DecodeElement(&text, &t); err != nil {
					return nil, fmt.Errorf("decode rate entry: %w", err)
				}
				text = strings.TrimSpace(text)
				var ratePtr *float64
				if text != "" {
					rate, err := strconv.ParseFloat(text, 64)
					if err != nil {
						return nil, fmt.Errorf("parse rate: %w", err)
					}
					ratePtr = floatPtr(rate)
				}

				var (
					age      int
					duration *int
				)

				if axisDepth > 1 {
					if !hasCurrentAge {
						return nil, fmt.Errorf("nested axis missing age identifier")
					}
					if !hasAttr {
						return nil, fmt.Errorf("nested axis missing duration identifier")
					}
					age = currentAge
					duration = intPtr(valueAttr)
				} else {
					if !hasAttr {
						return nil, fmt.Errorf("rate entry missing age identifier")
					}
					age = valueAttr
				}

				points = append(points, RatePoint{
					Table:    tableIndex,
					Age:      age,
					Duration: duration,
					Rate:     ratePtr,
				})
			}
		case xml.EndElement:
			name := strings.ToLower(t.Name.Local)
			switch name {
			case "axis":
				if inValues && axisDepth > 0 {
					if axisDepth == 1 {
						hasCurrentAge = false
					}
					axisDepth--
				}
			case "values":
				inValues = false
				axisDepth = 0
				hasCurrentAge = false
			}
		}
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no rate data found")
	}
	return points, nil
}

func attrInt(attrs []xml.Attr, name string) (int, bool) {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name.Local, name) {
			val, err := strconv.Atoi(strings.TrimSpace(attr.Value))
			if err != nil {
				return 0, false
			}
			return val, true
		}
	}
	return 0, false
}

func intPtr(v int) *int {
	p := new(int)
	*p = v
	return p
}

func floatPtr(v float64) *float64 {
	p := new(float64)
	*p = v
	return p
}
