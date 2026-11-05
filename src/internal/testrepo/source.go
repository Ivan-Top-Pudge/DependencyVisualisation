// Package testrepo читает тестовый граф зависимостей из локального JSON-файла.
package testrepo

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

var packageName = regexp.MustCompile(`^[A-Z]+$`)

// Source хранит тестовый граф с именами пакетов из заглавных латинских букв.
type Source struct {
	packages map[string][]string
}

// Load читает JSON-объект, сопоставляющий каждому пакету список его зависимостей.
// Все упомянутые зависимости должны быть объявлены ключами объекта.
func Load(path string) (*Source, error) {
	const op = "testrepo.Load"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var packages map[string][]string
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := validate(packages); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Source{packages: packages}, nil
}

// validate проверяет имена пакетов и наличие всех вершин, на которые ссылается граф.
func validate(packages map[string][]string) error {
	if packages == nil {
		return fmt.Errorf("test repository must be a JSON object")
	}
	for name, dependencies := range packages {
		if !packageName.MatchString(name) {
			return fmt.Errorf("invalid test package name %q", name)
		}
		for _, dependency := range dependencies {
			if !packageName.MatchString(dependency) {
				return fmt.Errorf("invalid dependency name %q for package %q", dependency, name)
			}
			if _, ok := packages[dependency]; !ok {
				return fmt.Errorf("dependency %q of package %q is not declared", dependency, name)
			}
		}
	}
	return nil
}

// LoadVersions возвращает единственную условную версию 1.0.0 тестового пакета.
// Требование к версии каждой зависимости равно *; неизвестное имя вызывает ошибку.
func (s *Source) LoadVersions(name string) ([]domain.PackageVersion, error) {
	const op = "testrepo.Source.LoadVersions"
	names, ok := s.packages[name]
	if !ok {
		return nil, fmt.Errorf("%s: package %q is not declared", op, name)
	}
	dependencies := make([]domain.Dependency, len(names))
	for i, dependency := range names {
		dependencies[i] = domain.Dependency{Name: dependency, Requirement: "*"}
	}
	return []domain.PackageVersion{{Name: name, Version: "1.0.0", Dependencies: dependencies}}, nil
}
