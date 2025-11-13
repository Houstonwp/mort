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
	parser := newRateParser()

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode rates: %w", err)
		}
		if err := parser.consume(dec, tok); err != nil {
			return nil, err
		}
	}

	return parser.result()
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

type rateParser struct {
	points        []RatePoint
	tableIndex    int
	inValues      bool
	axisDepth     int
	hasCurrentAge bool
	currentAge    int
}

func newRateParser() *rateParser {
	return &rateParser{tableIndex: -1}
}

func (rp *rateParser) consume(dec *xml.Decoder, tok xml.Token) error {
	switch t := tok.(type) {
	case xml.StartElement:
		name := strings.ToLower(t.Name.Local)
		switch name {
		case "table":
			rp.tableIndex++
		case "values":
			if rp.tableIndex >= 0 {
				rp.inValues = true
				rp.axisDepth = 0
				rp.hasCurrentAge = false
			}
		case "axis":
			if !rp.inValues {
				return nil
			}
			rp.axisDepth++
			if rp.axisDepth == 1 {
				if attr, ok := attrInt(t.Attr, "t"); ok {
					rp.currentAge = attr
					rp.hasCurrentAge = true
				} else {
					rp.hasCurrentAge = false
				}
			}
		case "y":
			if !rp.inValues {
				return nil
			}
			return rp.decodeRateEntry(dec, t)
		}
	case xml.EndElement:
		rp.handleEndElement(t)
	}
	return nil
}

func (rp *rateParser) decodeRateEntry(dec *xml.Decoder, start xml.StartElement) error {
	valueAttr, hasAttr := attrInt(start.Attr, "t")
	var text string
	if err := dec.DecodeElement(&text, &start); err != nil {
		return fmt.Errorf("decode rate entry: %w", err)
	}
	text = strings.TrimSpace(text)
	var ratePtr *float64
	if text != "" {
		rate, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return fmt.Errorf("parse rate: %w", err)
		}
		ratePtr = floatPtr(rate)
	}

	var (
		age      int
		duration *int
	)

	if rp.axisDepth > 1 {
		if !rp.hasCurrentAge {
			return fmt.Errorf("nested axis missing age identifier")
		}
		if !hasAttr {
			return fmt.Errorf("nested axis missing duration identifier")
		}
		age = rp.currentAge
		duration = intPtr(valueAttr)
	} else {
		if !hasAttr {
			return fmt.Errorf("rate entry missing age identifier")
		}
		age = valueAttr
	}

	rp.points = append(rp.points, RatePoint{
		Table:    rp.tableIndex,
		Age:      age,
		Duration: duration,
		Rate:     ratePtr,
	})
	return nil
}

func (rp *rateParser) handleEndElement(end xml.EndElement) {
	name := strings.ToLower(end.Name.Local)
	switch name {
	case "axis":
		if rp.inValues && rp.axisDepth > 0 {
			if rp.axisDepth == 1 {
				rp.hasCurrentAge = false
			}
			rp.axisDepth--
		}
	case "values":
		rp.inValues = false
		rp.axisDepth = 0
		rp.hasCurrentAge = false
	}
}

func (rp *rateParser) result() ([]RatePoint, error) {
	if len(rp.points) == 0 {
		return nil, fmt.Errorf("no rate data found")
	}
	return rp.points, nil
}
