package cargo

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// TestParseVersions проверяет поля версий и настоящие имена зависимостей из JSONL.
func TestParseVersions(t *testing.T) {
	file, err := os.Open("../../../tests/cargo_versions.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	got, err := parse(file)
	if err != nil {
		t.Fatalf("parse() failed: %v", err)
	}
	if want := expectedPackageVersions(); !reflect.DeepEqual(got, want) {
		t.Errorf("parse() = %+v, want %+v", got, want)
	}
}

// expectedPackageVersions задает ожидаемые данные из tests/cargo_versions.jsonl.
func expectedPackageVersions() []domain.PackageVersion {
	return []domain.PackageVersion{
		{Name: "demo", Version: "0.1.0", Yanked: true, Dependencies: []domain.Dependency{}},
		{Name: "demo", Version: "0.2.0", Dependencies: []domain.Dependency{
			{Name: "serde", Requirement: "^1", PackageName: "serde"},
		}},
		{Name: "demo", Version: "0.3.0", Dependencies: []domain.Dependency{
			{Name: "log", Requirement: "^0.4"},
		}},
		{Name: "demo", Version: "0.4.0", Dependencies: []domain.Dependency{
			{Name: "itoa", Requirement: "^1"},
		}},
	}
}

// TestParseInvalidJSON проверяет ошибки формата и отсутствие частичного результата.
func TestParseInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid_json", "not json"},
		{"truncated_object", `{"name":"demo","vers":"0.1.0","deps":[`},
		{"invalid_after_version", "{\"name\":\"demo\",\"vers\":\"0.1.0\",\"deps\":[]}\n{\"name\":"},
		{"invalid_dependencies", `{"name":"demo","vers":"0.1.0","deps":"serde"}`},
		{"array_instead_of_object", `[{"name":"demo","vers":"0.1.0","deps":[]}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := parse(strings.NewReader(tt.data))
			if err == nil {
				t.Fatal("parse() succeeded for invalid JSONL")
			}
			if versions != nil {
				t.Errorf("parse() returned a partial result: %+v", versions)
			}
		})
	}
}

// TestParseEmptyInput проверяет пустой поток и поток из пробельных символов.
func TestParseEmptyInput(t *testing.T) {
	const emptyVersionCount = 0
	for _, data := range []string{"", " \n\t "} {
		versions, err := parse(strings.NewReader(data))
		if err != nil {
			t.Fatalf("parse() failed for empty input: %v", err)
		}
		if len(versions) != emptyVersionCount {
			t.Errorf("parse() returned versions for empty input: %+v", versions)
		}
	}
}

// TestParseReadError проверяет передачу исходной ошибки чтения вызывающему коду.
func TestParseReadError(t *testing.T) {
	readErr := errors.New("reader failed")
	versions, err := parse(iotest.ErrReader(readErr))
	if !errors.Is(err, readErr) {
		t.Fatalf("parse() error = %v, want wrapped %v", err, readErr)
	}
	if versions != nil {
		t.Errorf("parse() returned versions after a read error: %+v", versions)
	}
}
