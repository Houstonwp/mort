package xtbml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type documentData struct {
	version        string
	classification *ContentClassification
	tableMetas     []TableMeta
	rates          []RatePoint
}

func parseDocument(data []byte) (*documentData, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	parser := newRateParser()
	doc := &documentData{version: "unknown"}
	tableIndex := -1

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode xtbml: %w", err)
		}
		if err := parser.consume(dec, tok); err != nil {
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch strings.ToLower(start.Name.Local) {
		case "xtbml":
			if doc.version == "unknown" {
				if v := versionFromAttrs(start.Attr); v != "" {
					doc.version = v
				}
			}
		case "table":
			tableIndex++
			doc.ensureMetaCapacity(tableIndex + 1)
		case "contentclassification":
			if doc.classification != nil {
				if err := dec.Skip(); err != nil {
					return nil, fmt.Errorf("skip duplicate content classification: %w", err)
				}
				continue
			}
			classification, err := decodeClassificationElement(dec, start)
			if err != nil {
				return nil, err
			}
			doc.classification = classification
		case "metadata":
			meta, err := decodeMetaElement(dec, start)
			if err != nil {
				return nil, err
			}
			if tableIndex >= 0 && tableIndex < len(doc.tableMetas) {
				doc.tableMetas[tableIndex] = meta
			}
		}
	}

	rates, err := parser.result()
	if err != nil {
		return nil, err
	}
	doc.rates = rates

	if doc.classification == nil {
		return nil, fmt.Errorf("content classification missing table name")
	}

	return doc, nil
}

func (d *documentData) ensureMetaCapacity(size int) {
	for len(d.tableMetas) < size {
		d.tableMetas = append(d.tableMetas, TableMeta{})
	}
}

func versionFromAttrs(attrs []xml.Attr) string {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Name.Local, "version") && attr.Value != "" {
			return attr.Value
		}
	}
	return ""
}
