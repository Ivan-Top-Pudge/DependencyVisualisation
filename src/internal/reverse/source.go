// Package reverse строит индекс входящих зависимостей для обхода репозитория в ширину.
package reverse

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// Source хранит выбранные релизы и имена пакетов, зависящих от каждого из них.
type Source struct {
	releases   map[string]domain.PackageID
	dependents map[string][]string
}

// New создает пустой индекс для заполнения методом AddVersions.
func New() *Source {
	return &Source{releases: make(map[string]domain.PackageID), dependents: make(map[string][]string)}
}

// AddVersions добавляет входящие связи выбранного релиза одного пакета из снимка.
// Пакеты без стабильного релиза и зависимости из сторонних реестров пропускаются.
func (s *Source) AddVersions(versions []domain.PackageVersion) error {
	pkg, err := analyzer.SelectVersion(versions)
	if errors.Is(err, analyzer.ErrNoRelease) {
		return nil
	}
	if err != nil {
		return err
	}
	key := strings.ToLower(pkg.Name)
	if _, exists := s.releases[key]; exists {
		return fmt.Errorf("duplicate package %q in repository", pkg.Name)
	}
	s.releases[key] = domain.PackageID{Name: pkg.Name, Version: pkg.Version}
	seen := make(map[string]bool)
	for _, dep := range pkg.Dependencies {
		target := strings.ToLower(dep.Name)
		if dep.Registry == "" && !seen[target] {
			s.dependents[target] = append(s.dependents[target], pkg.Name)
			seen[target] = true
		}
	}
	return nil
}

// LoadVersions возвращает выбранный релиз с отсортированным списком прямых обратных зависимостей.
func (s *Source) LoadVersions(name string) ([]domain.PackageVersion, error) {
	key := strings.ToLower(name)
	id, exists := s.releases[key]
	if !exists {
		return nil, fmt.Errorf("package %q has no stable non-yanked release in repository", name)
	}
	names := s.dependents[key]
	deps := make([]domain.Dependency, len(names))
	for i, dependent := range names {
		deps[i].Name = dependent
	}
	slices.SortFunc(deps, func(a, b domain.Dependency) int { return strings.Compare(a.Name, b.Name) })
	pkg := domain.PackageVersion{Name: id.Name, Version: id.Version, Dependencies: deps}
	return []domain.PackageVersion{pkg}, nil
}
