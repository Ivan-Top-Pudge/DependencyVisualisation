package analyzer_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/cargo"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// TestPackageDependencies проверяет публичный конструктор и выбор пакета после HTTP-загрузки.
func TestPackageDependencies(t *testing.T) {
	data, err := os.ReadFile("../../../tests/cargo_versions.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	const skippedVersions = `{"name":"demo","vers":"9.0.0","deps":[],"yanked":true}
{"name":"demo","vers":"10.0.0-beta","deps":[],"yanked":false}`
	server := newAnalyzerServer(t, http.StatusOK, string(data)+skippedVersions)
	got, err := analyzer.New(cargo.NewClient(server.URL)).PackageDependencies("demo")
	if err != nil {
		t.Fatalf("PackageDependencies() failed: %v", err)
	}
	want := domain.PackageVersion{
		Name: "demo", Version: "0.4.0",
		Dependencies: []domain.Dependency{{Name: "itoa", Requirement: "^1"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PackageDependencies() = %+v, want %+v", got, want)
	}
}

// TestPackageDependenciesHTTPError проверяет передачу ошибки загрузки с контекстом анализатора.
func TestPackageDependenciesHTTPError(t *testing.T) {
	server := newAnalyzerServer(t, http.StatusNotFound, "not JSON")
	got, err := analyzer.New(cargo.NewClient(server.URL)).PackageDependencies("demo")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("PackageDependencies() error = %v, want HTTP 404", err)
	}
	if !strings.Contains(err.Error(), "analyzer.Analyzer.PackageDependencies") {
		t.Errorf("error is missing analyzer context: %v", err)
	}
	if !reflect.DeepEqual(got, domain.PackageVersion{}) {
		t.Errorf("PackageDependencies() returned a package after an HTTP error: %+v", got)
	}
}

// TestPackageDependenciesInvalidJSON проверяет сохранение типа ошибки разбора ответа.
func TestPackageDependenciesInvalidJSON(t *testing.T) {
	server := newAnalyzerServer(t, http.StatusOK, "not JSON")
	got, err := analyzer.New(cargo.NewClient(server.URL)).PackageDependencies("demo")
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("PackageDependencies() error = %v, want a JSON syntax error", err)
	}
	if !reflect.DeepEqual(got, domain.PackageVersion{}) {
		t.Errorf("PackageDependencies() returned a package after a parse error: %+v", got)
	}
}

// TestPackageDependenciesNoRelease проверяет ошибку выбора для пустого и неподходящего индекса.
func TestPackageDependenciesNoRelease(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty", ""},
		{"yanked", `{"name":"demo","vers":"1.0.0","deps":[],"yanked":true}`},
		{"prerelease", `{"name":"demo","vers":"2.0.0-beta","deps":[],"yanked":false}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newAnalyzerServer(t, http.StatusOK, tt.body)
			got, err := analyzer.New(cargo.NewClient(server.URL)).PackageDependencies("demo")
			if err == nil || !strings.Contains(err.Error(), "no non-yanked release versions") {
				t.Fatalf("PackageDependencies() error = %v, want no suitable release", err)
			}
			if !reflect.DeepEqual(got, domain.PackageVersion{}) {
				t.Errorf("PackageDependencies() returned a package after a selection error: %+v", got)
			}
		})
	}
}

// TestPackageDependenciesInvalidVersion проверяет передачу ошибки номера загруженной версии.
func TestPackageDependenciesInvalidVersion(t *testing.T) {
	const body = `{"name":"demo","vers":"invalid","deps":[],"yanked":false}`
	server := newAnalyzerServer(t, http.StatusOK, body)
	got, err := analyzer.New(cargo.NewClient(server.URL)).PackageDependencies("demo")
	if err == nil || !strings.Contains(err.Error(), "invalid semantic version") {
		t.Fatalf("PackageDependencies() error = %v, want an invalid version error", err)
	}
	if !reflect.DeepEqual(got, domain.PackageVersion{}) {
		t.Errorf("PackageDependencies() returned a package after an invalid version: %+v", got)
	}
}

// TestAnalyzerNewInvalidURL проверяет отложенную проверку URL в публичном конструкторе.
func TestAnalyzerNewInvalidURL(t *testing.T) {
	client := analyzer.New(cargo.NewClient("not-a-url"))
	if client == nil {
		t.Fatal("New() returned nil")
	}
	got, err := client.PackageDependencies("demo")
	if err == nil {
		t.Fatal("PackageDependencies() accepted an invalid repository URL")
	}
	if !reflect.DeepEqual(got, domain.PackageVersion{}) {
		t.Errorf("PackageDependencies() returned a package for an invalid URL: %+v", got)
	}
}

// newAnalyzerServer создает локальный индекс с проверкой метода и пути запроса пакета demo.
func newAnalyzerServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/de/mo/demo" {
			t.Errorf("request path = %q, want /de/mo/demo", r.URL.Path)
		}
		w.WriteHeader(status)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
