package reverse

import (
	"reflect"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// TestIncomingEdges проверяет повтор ребра, порядок соседей и сохранность исходных данных.
func TestIncomingEdges(t *testing.T) {
	source := New()
	root := domain.PackageVersion{Name: "A", Version: "1.0.0"}
	deps := []domain.Dependency{{Name: "A"}, {Name: "A"}}
	for _, name := range []string{"C", "B"} {
		pkg := domain.PackageVersion{Name: name, Version: "2.0.0", Dependencies: deps}
		if err := source.AddVersions([]domain.PackageVersion{pkg}); err != nil {
			t.Fatal(err)
		}
	}
	if err := source.AddVersions([]domain.PackageVersion{root}); err != nil {
		t.Fatal(err)
	}
	got, err := source.LoadVersions("A")
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.PackageVersion{{Name: "A", Version: "1.0.0", Dependencies: []domain.Dependency{
		{Name: "B"}, {Name: "C"},
	}}}
	if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(deps, []domain.Dependency{{Name: "A"}, {Name: "A"}}) {
		t.Errorf("обратные связи %+v, исходные зависимости %+v", got, deps)
	}
	if err := source.AddVersions([]domain.PackageVersion{root}); err == nil {
		t.Error("принята повторная запись пакета")
	}
}
