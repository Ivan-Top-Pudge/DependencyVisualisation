package testrepo

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// TestSource проверяет преобразование графа в версии, конечную вершину и неизвестное имя.
func TestSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "graph.json")
	if err := os.WriteFile(path, []byte(`{"A":["B","A"],"B":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	source, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := source.LoadVersions("A")
	want := []domain.PackageVersion{{Name: "A", Version: "1.0.0", Dependencies: []domain.Dependency{
		{Name: "B", Requirement: "*"}, {Name: "A", Requirement: "*"},
	}}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadVersions(A) = %+v, %v, want %+v", got, err, want)
	}
	leaf, err := source.LoadVersions("B")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(leaf, []domain.PackageVersion{{
		Name: "B", Version: "1.0.0", Dependencies: []domain.Dependency{},
	}}) {
		t.Errorf("LoadVersions(B) = %+v, want a leaf", leaf)
	}
	if _, err := source.LoadVersions("Z"); err == nil {
		t.Error("LoadVersions(Z) accepted an unknown package")
	}
}

// TestLoadErrors проверяет ошибки файла, JSON, имен и необъявленных зависимостей.
func TestLoadErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "graph.json")
	if _, err := Load(path); err == nil {
		t.Error("Load() accepted a missing file")
	}
	tests := []struct {
		name string
		body string
		want string
	}{
		{"json", `{`, "unexpected end"},
		{"null", `null`, "JSON object"},
		{"package", `{"a":[]}`, "invalid test package name"},
		{"dependency", `{"A":["b"]}`, "invalid dependency name"},
		{"missing", `{"A":["B"]}`, "not declared"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tt.body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Load() error = %v, want %q", err, tt.want)
			}
		})
	}
}
