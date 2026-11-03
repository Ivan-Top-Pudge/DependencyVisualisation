package cargo

import "testing"

// TestBuildURL проверяет адреса индекса Cargo для разных имен и базовых URL.
func TestBuildURL(t *testing.T) {
	const repositoryURL = "https://index.crates.io"
	tests := []struct {
		name        string
		repository  string
		packageName string
		want        string
	}{
		{"one_character", repositoryURL, "A", repositoryURL + "/1/a"},
		{"two_characters", repositoryURL, "AB", repositoryURL + "/2/ab"},
		{"three_characters", repositoryURL, "ABC", repositoryURL + "/3/a/abc"},
		{"four_characters", repositoryURL, "ABCD", repositoryURL + "/ab/cd/abcd"},
		{"lowercase", repositoryURL, "Serde", repositoryURL + "/se/rd/serde"},
		{"underscore", repositoryURL, "serde_json", repositoryURL + "/se/rd/serde_json"},
		{"hyphen", repositoryURL, "proc-macro2", repositoryURL + "/pr/oc/proc-macro2"},
		{"trailing_slash", repositoryURL + "/", "serde", repositoryURL + "/se/rd/serde"},
		{"base_path", repositoryURL + "/registry/", "serde", repositoryURL + "/registry/se/rd/serde"},
		{"query", repositoryURL + "?key=value", "serde", repositoryURL + "/se/rd/serde?key=value"},
		{"unicode_three_characters", repositoryURL, "AЯB", repositoryURL + "/3/a/a%D1%8Fb"},
		{"unicode_prefix", repositoryURL, "ЯABC", repositoryURL + "/%D1%8Fa/bc/%D1%8Fabc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.repository).buildURL(tt.packageName)
			if err != nil {
				t.Fatalf("buildURL() failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("buildURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBuildURLInvalidInput проверяет отклонение неверных имен и URL репозитория.
func TestBuildURLInvalidInput(t *testing.T) {
	const repositoryURL = "https://index.crates.io"
	tests := []struct {
		name        string
		repository  string
		packageName string
	}{
		{"empty_name", repositoryURL, ""},
		{"slash", repositoryURL, "serde/json"},
		{"space", repositoryURL, "serde json"},
		{"parent_path", repositoryURL, "../serde"},
		{"query_in_name", repositoryURL, "serde?key=value"},
		{"empty_repository", "", "serde"},
		{"relative_repository", "/registry", "serde"},
		{"unsupported_scheme", "ftp://index.crates.io", "serde"},
		{"missing_host", "https:///registry", "serde"},
		{"invalid_escape", "https://index.crates.io/%", "serde"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.repository).buildURL(tt.packageName)
			if err == nil {
				t.Fatal("buildURL() succeeded for invalid input")
			}
			if got != "" {
				t.Errorf("buildURL() returned %q with an error, want empty URL", got)
			}
		})
	}
}
