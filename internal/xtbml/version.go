package xtbml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// InferVersion reads the XTbml root and returns its version attribute.
func InferVersion(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("decode xtbml: %w", err)
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !strings.EqualFold(start.Name.Local, "XTbml") {
			continue
		}

		for _, attr := range start.Attr {
			if attr.Name.Local == "version" && attr.Value != "" {
				return attr.Value, nil
			}
		}
		return "unknown", nil
	}

	return "unknown", nil
}
