package xtbml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseContentClassification(t *testing.T) {
	path := filepath.Join("testdata", "classification_sample.xml")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	class, err := ParseContentClassification(f)
	if err != nil {
		t.Fatalf("ParseContentClassification() error = %v", err)
	}

	if class.TableIdentity != "1941" || class.ProviderDomain != "soa.org" {
		t.Fatalf("unexpected identity/domain: %#v", class)
	}
	if class.ProviderName != "Society of Actuaries" || !strings.Contains(class.TableReference, "SOA") {
		t.Fatalf("provider/ref mismatch: %#v", class)
	}
	if class.ContentType.Code != "85" || class.ContentType.Label != "CSO / CET" {
		t.Fatalf("content type mismatch: %#v", class.ContentType)
	}
	if len(class.Keywords) != 2 || class.Keywords[0] != "Aggregate" {
		t.Fatalf("keywords mismatch: %#v", class.Keywords)
	}
}
