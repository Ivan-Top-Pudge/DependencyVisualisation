package analyzer

import (
	"cmp"
	"testing"
)

// TestCompareVersions проверяет основные компоненты, суффиксы и большие числовые значения.
func TestCompareVersions(t *testing.T) {
	tests := []struct {
		left, right string
		want        int
	}{
		{"1.2.3", "1.2.3", 0},
		{"0.0.0", "0.0.0", 0},
		{"2.0.0", "1.99.99", 1},
		{"1.10.0", "1.9.9", 1},
		{"1.0.10", "1.0.9", 1},
		{"0.9.0", "1.0.0", -1},
		{"1.2.3+one", "1.2.3+two", 0},
		{"1.0.0+001", "1.0.0+1", 0},
		{"1.0.0+build-42", "1.0.0", 0},
		{"1.0.0", "1.0.0-rc.1", 1},
		{"1.0.0-beta.11", "1.0.0-beta.2", 1},
		{"1.0.0-1", "1.0.0-alpha", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha", 1},
		{"1.0.0-A", "1.0.0-a", -1},
		{"1.0.0-01a", "1.0.0-1a", -1},
		{"2.0.0-alpha", "1.99.99", 1},
		{"18446744073709551616.0.0", "18446744073709551615.0.0", 1},
		{"1.0.0-18446744073709551616", "1.0.0-18446744073709551615", 1},
		{"1.0.0--", "1.0.0-alpha", -1},
	}
	for _, tt := range tests {
		t.Run(tt.left+"_vs_"+tt.right, func(t *testing.T) {
			got, err := CompareVersions(tt.left, tt.right)
			if err != nil {
				t.Fatalf("CompareVersions() failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("CompareVersions() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestCompareVersionsSemVerOrder проверяет все пары из примера порядка SemVer 2.0.0.
func TestCompareVersionsSemVerOrder(t *testing.T) {
	ordered := []string{
		"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta",
		"1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0",
	}
	for i, left := range ordered {
		for j, right := range ordered {
			got, err := CompareVersions(left, right)
			if err != nil {
				t.Fatalf("CompareVersions(%q, %q) failed: %v", left, right, err)
			}
			if want := cmp.Compare(i, j); got != want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", left, right, got, want)
			}
		}
	}
}

// TestCompareVersionsInvalidInput проверяет формат обеих строк версии независимо.
func TestCompareVersionsInvalidInput(t *testing.T) {
	invalid := []string{
		"", "v1.2.3", "1", "1.2", "1.2.3.4", "01.2.3", "1.02.3", "1.2.03",
		"-1.2.3", "1.-2.3", "1.2.-3", "1.2.a", "1.0.0-01", "1.0.0-alpha.01",
		"1.0.0-", "1.0.0-alpha..1", "1.0.0-α", "1.0.0+", "1.0.0+one..two",
		"1.0.0+one+two", "1.0.0+a_b", " 1.2.3", "1.2.3 ", "1.2.3\n",
	}
	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			pairs := []struct{ left, right string }{{value, "1.0.0"}, {"1.0.0", value}}
			for _, pair := range pairs {
				if _, err := CompareVersions(pair.left, pair.right); err == nil {
					t.Errorf("CompareVersions(%q, %q) succeeded for invalid input", pair.left, pair.right)
				}
			}
		})
	}
}
