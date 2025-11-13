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
	var doc struct {
		ContentClassification struct {
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
		} `xml:"ContentClassification"`
	}

	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode content classification: %w", err)
	}

	cc := doc.ContentClassification
	if cc.TableName == "" {
		return nil, fmt.Errorf("content classification missing table name")
	}

	return &ContentClassification{
		TableIdentity:  strings.TrimSpace(cc.TableIdentity),
		ProviderDomain: strings.TrimSpace(cc.ProviderDomain),
		ProviderName:   strings.TrimSpace(cc.ProviderName),
		TableReference: strings.TrimSpace(cc.TableReference),
		ContentType: ClassifiedValue{
			Code:  strings.TrimSpace(cc.ContentType.Code),
			Label: strings.TrimSpace(cc.ContentType.Label),
		},
		TableName:        strings.TrimSpace(cc.TableName),
		TableDescription: strings.TrimSpace(cc.TableDescription),
		Comments:         strings.TrimSpace(cc.Comments),
		Keywords:         normalizeKeywords(cc.Keywords),
	}, nil
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
