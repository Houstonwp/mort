package xtbml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRates(t *testing.T) {
	t.Run("reads single-axis values", func(t *testing.T) {
		path := filepath.Join("testdata", "rates_simple.xml")
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer f.Close()

		points, err := ParseRates(f)
		if err != nil {
			t.Fatalf("ParseRates() err = %v", err)
		}
		if len(points) != 2 {
			t.Fatalf("len(points) = %d, want 2", len(points))
		}
		if points[0].Table != 0 || points[0].Age != 40 || points[0].Duration != nil || points[0].Rate == nil || *points[0].Rate != 0.01 {
			t.Fatalf("point[0] mismatch: %#v", points[0])
		}
		if points[1].Age != 41 || points[1].Rate == nil || *points[1].Rate != 0.011 {
			t.Fatalf("point[1] mismatch: %#v", points[1])
		}
	})

	t.Run("reads nested axes and multiple tables", func(t *testing.T) {
		path := filepath.Join("testdata", "rates_nested.xml")
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer f.Close()

		points, err := ParseRates(f)
		if err != nil {
			t.Fatalf("ParseRates() err = %v", err)
		}
		if len(points) != 4 {
			t.Fatalf("len(points) = %d, want 4", len(points))
		}
		if points[0].Table != 0 || points[0].Age != 40 || points[0].Duration == nil || *points[0].Duration != 0 || points[0].Rate == nil {
			t.Fatalf("nested point[0] mismatch: %#v", points[0])
		}
		if points[1].Age != 40 || points[1].Duration == nil || *points[1].Duration != 1 || points[1].Rate == nil {
			t.Fatalf("nested point[1] mismatch: %#v", points[1])
		}
		if points[2].Age != 41 || points[2].Duration == nil || points[2].Rate == nil {
			t.Fatalf("nested point[2] mismatch: %#v", points[2])
		}
		if points[3].Table != 1 || points[3].Age != 90 || points[3].Duration != nil || points[3].Rate == nil {
			t.Fatalf("second table point mismatch: %#v", points[3])
		}
	})

	t.Run("empty values allowed", func(t *testing.T) {
		xml := `<XTbML><Table><Values><Axis><Y t="10"></Y><Y t="11">0.5</Y></Axis></Values></Table></XTbML>`
		points, err := ParseRates(strings.NewReader(xml))
		if err != nil {
			t.Fatalf("ParseRates() err = %v", err)
		}
		if points[0].Rate != nil {
			t.Fatalf("expected nil rate for empty value: %#v", points[0])
		}
		if points[1].Rate == nil || *points[1].Rate != 0.5 {
			t.Fatalf("expected parsed rate for second value: %#v", points[1])
		}
	})

	t.Run("missing identifiers error", func(t *testing.T) {
		xml := `<XTbML><Table><Values><Axis t="1"><Axis><Y>0.1</Y></Axis></Axis></Values></Table></XTbML>`
		_, err := ParseRates(strings.NewReader(xml))
		if err == nil {
			t.Fatal("ParseRates() expected error, got nil")
		}
	})
}
