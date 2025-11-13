package xtbml

import "testing"

func TestNormalizeIdentifier(t *testing.T) {
	t.Run("collapses whitespace and punctuation", func(t *testing.T) {
		got := NormalizeIdentifier("2012 IAM Basic-Table")
		want := "2012_iam_basic_table"
		if got != want {
			t.Fatalf("NormalizeIdentifier() = %q, want %q", got, want)
		}
	})

	t.Run("trims leading/trailing separators", func(t *testing.T) {
		got := NormalizeIdentifier("  --RP-2000-- ")
		want := "rp_2000"
		if got != want {
			t.Fatalf("NormalizeIdentifier() = %q, want %q", got, want)
		}
	})
}
