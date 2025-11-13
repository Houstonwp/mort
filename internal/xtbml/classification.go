package xtbml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ClassifiedValue captures a text label that also has a tc attribute.
type ClassifiedValue struct {
	Code  string
	Label string
}

// ContentClassification holds the document-level metadata.
type ContentClassification struct {
	TableIdentity    string
	ProviderDomain   string
	ProviderName     string
	TableReference   string
	ContentType      ClassifiedValue
	TableName        string
	TableDescription string
	Comments         string
	Keywords         []string
}

// ParseContentClassification extracts the <ContentClassification> block.
func ParseContentClassification(r io.Reader) (*ContentClassification, error) {
	dec := xml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("content classification missing table name")
			}
			return nil, fmt.Errorf("decode content classification: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok || !strings.EqualFold(start.Name.Local, "contentclassification") {
			continue
		}
		return decodeClassificationElement(dec, start)
	}
}

func normalizeKeywords(list []string) []string {
	out := make([]string, 0, len(list))
	for _, kw := range list {
		trim := strings.TrimSpace(kw)
		if trim != "" {
			out = append(out, trim)
		}
	}
	return out
}

func decodeClassificationElement(dec *xml.Decoder, start xml.StartElement) (*ContentClassification, error) {
	var node struct {
		TableIdentity  string `xml:"TableIdentity"`
		ProviderDomain string `xml:"ProviderDomain"`
		ProviderName   string `xml:"ProviderName"`
		TableReference string `xml:"TableReference"`
		ContentType    struct {
			Code  string `xml:"tc,attr"`
			Label string `xml:",chardata"`
		} `xml:"ContentType"`
		TableName        string   `xml:"TableName"`
		TableDescription string   `xml:"TableDescription"`
		Comments         string   `xml:"Comments"`
		Keywords         []string `xml:"KeyWord"`
	}

	if err := dec.DecodeElement(&node, &start); err != nil {
		return nil, fmt.Errorf("decode content classification: %w", err)
	}
	if strings.TrimSpace(node.TableName) == "" {
		return nil, fmt.Errorf("content classification missing table name")
	}
	return &ContentClassification{
		TableIdentity:  strings.TrimSpace(node.TableIdentity),
		ProviderDomain: strings.TrimSpace(node.ProviderDomain),
		ProviderName:   strings.TrimSpace(node.ProviderName),
		TableReference: strings.TrimSpace(node.TableReference),
		ContentType: ClassifiedValue{
			Code:  strings.TrimSpace(node.ContentType.Code),
			Label: strings.TrimSpace(node.ContentType.Label),
		},
		TableName:        strings.TrimSpace(node.TableName),
		TableDescription: strings.TrimSpace(node.TableDescription),
		Comments:         strings.TrimSpace(node.Comments),
		Keywords:         normalizeKeywords(node.Keywords),
	}, nil
}
