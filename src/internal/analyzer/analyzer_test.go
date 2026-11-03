package analyzer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// versionSelectionCase задает список версий и ожидаемый номер выбранного релиза.
type versionSelectionCase struct {
	name     string
	versions []domain.PackageVersion
	want     string
}

// TestSelectVersion проверяет максимум, фильтрацию отозванных и предварительных версий.
func TestSelectVersion(t *testing.T) {
	for _, tt := range validVersionSelectionCases() {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&Analyzer{}).selectVersion(tt.versions)
			if err != nil {
				t.Fatalf("selectVersion() failed: %v", err)
			}
			if got.Version != tt.want {
				t.Errorf("selectVersion() version = %q, want %q", got.Version, tt.want)
			}
		})
	}
}

// validVersionSelectionCases задает независимые примеры успешного выбора версии.
func validVersionSelectionCases() []versionSelectionCase {
	return []versionSelectionCase{
		{"unordered", []domain.PackageVersion{
			{Version: "1.9.0"}, {Version: "1.10.0"}, {Version: "1.2.0"},
		}, "1.10.0"},
		{"first_yanked", []domain.PackageVersion{
			{Version: "2.0.0", Yanked: true}, {Version: "1.0.0"},
		}, "1.0.0"},
		{"later_yanked", []domain.PackageVersion{
			{Version: "1.0.0"}, {Version: "2.0.0", Yanked: true},
		}, "1.0.0"},
		{"prerelease", []domain.PackageVersion{
			{Version: "2.0.0-rc.1"}, {Version: "1.0.0"}, {Version: "3.0.0-alpha"},
		}, "1.0.0"},
		{"hyphen_in_build", []domain.PackageVersion{
			{Version: "1.0.0+build-42"}, {Version: "0.9.0"},
		}, "1.0.0+build-42"},
		{"zero_version", []domain.PackageVersion{{Version: "0.0.0"}}, "0.0.0"},
		{"invalid_yanked", []domain.PackageVersion{
			{Version: "invalid", Yanked: true}, {Version: "1.0.0"},
		}, "1.0.0"},
	}
}

// TestSelectVersionPreservesPackage проверяет данные пакета и порядок входного среза.
func TestSelectVersionPreservesPackage(t *testing.T) {
	want := domain.PackageVersion{
		Name: "demo", Version: "1.10.0",
		Dependencies: []domain.Dependency{{Name: "serde", Requirement: "^1", PackageName: "serde"}},
	}
	versions := []domain.PackageVersion{
		{Name: "demo", Version: "2.0.0-beta"}, want, {Name: "demo", Version: "1.9.0"},
	}
	before := append([]domain.PackageVersion(nil), versions...)
	got, err := (&Analyzer{}).selectVersion(versions)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("selectVersion() = %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(versions, before) {
		t.Errorf("selectVersion() changed input: %+v, want %+v", versions, before)
	}
}

// TestSelectVersionEqualPriority проверяет сохранение первого из равных по приоритету релизов.
func TestSelectVersionEqualPriority(t *testing.T) {
	first := domain.PackageVersion{Name: "demo", Version: "1.0.0+first"}
	second := domain.PackageVersion{Name: "demo", Version: "1.0.0+second"}
	got, err := (&Analyzer{}).selectVersion([]domain.PackageVersion{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, first) {
		t.Errorf("selectVersion() = %+v, want first release %+v", got, first)
	}
}

// TestSelectVersionErrors проверяет отсутствие подходящих версий и неверные номера.
func TestSelectVersionErrors(t *testing.T) {
	const noVersion = "no non-yanked release versions"
	const badVersion = "invalid semantic version"
	tests := []struct {
		name      string
		versions  []domain.PackageVersion
		wantError string
	}{
		{"nil", nil, noVersion},
		{"empty", []domain.PackageVersion{}, noVersion},
		{"all_yanked", []domain.PackageVersion{{Version: "1.0.0", Yanked: true}}, noVersion},
		{"all_prerelease", []domain.PackageVersion{{Version: "2.0.0-beta"}}, noVersion},
		{"mixed_unavailable", []domain.PackageVersion{
			{Version: "1.0.0", Yanked: true}, {Version: "2.0.0-beta"},
		}, noVersion},
		{"invalid_first", []domain.PackageVersion{{Version: "invalid"}}, badVersion},
		{"invalid_after_release", []domain.PackageVersion{
			{Version: "1.0.0"}, {Version: "invalid"},
		}, badVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&Analyzer{}).selectVersion(tt.versions)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("selectVersion() error = %v, want %q", err, tt.wantError)
			}
			if !reflect.DeepEqual(got, domain.PackageVersion{}) {
				t.Errorf("selectVersion() returned a package after an error: %+v", got)
			}
		})
	}
}
