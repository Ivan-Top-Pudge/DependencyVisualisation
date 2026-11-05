package analyzer

import (
	"fmt"
	"slices"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// PackageSource загружает версии пакета из индекса Cargo или тестового файла.
type PackageSource interface {
	LoadVersions(name string) ([]domain.PackageVersion, error)
}

// Analyzer объединяет операции анализа версий пакетов Cargo.
type Analyzer struct {
	source PackageSource
}

// New создает анализатор для выбранного источника пакетов.
func New(source PackageSource) *Analyzer {
	return &Analyzer{
		source: source,
	}
}

// PackageDependencies загружает версии пакета и возвращает выбранную с прямыми зависимостями.
// Выбирается максимальная по SemVer неотозванная версия без prerelease.
// Ошибки загрузки, разбора и выбора версии возвращаются с контекстом операции.
func (a *Analyzer) PackageDependencies(name string) (domain.PackageVersion, error) {
	const op = "analyzer.Analyzer.PackageDependencies"
	versions, err := a.source.LoadVersions(name)
	if err != nil {
		return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
	}

	selectedVersion, err := a.selectVersion(versions)
	if err != nil {
		return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
	}

	return selectedVersion, nil
}

// BuildGraph строит граф обходом BFS без рекурсии, считая глубину корня равной нулю.
// Вершины на maxDepth включаются без раскрытия зависимостей; циклические ребра сохраняются.
// Каждый пакет загружается один раз; требования к версиям зависимостей не учитываются.
func (a *Analyzer) BuildGraph(name string, maxDepth int) (domain.Graph, error) {
	const op = "analyzer.Analyzer.BuildGraph"
	const rootDepth, emptyQueue = 0, 0
	if maxDepth < rootDepth {
		return domain.Graph{}, fmt.Errorf("%s: maxDepth must be non-negative", op)
	}
	root, err := a.PackageDependencies(name)
	if err != nil {
		return domain.Graph{}, fmt.Errorf("%s: %w", op, err)
	}

	rootID := domain.PackageID{Name: root.Name, Version: root.Version}
	graph := domain.Graph{
		Root:  rootID,
		Nodes: map[domain.PackageID]domain.PackageVersion{rootID: root},
		Edges: make(map[domain.PackageID][]domain.PackageID),
	}
	packages := map[string]domain.PackageVersion{name: root, root.Name: root}
	queue := []domain.PackageID{rootID}
	for depth := rootDepth; depth < maxDepth && len(queue) != emptyQueue; depth++ {
		var next []domain.PackageID
		for _, id := range queue {
			discovered, err := a.expandGraph(&graph, id, packages)
			if err != nil {
				return domain.Graph{}, fmt.Errorf("%s: %w", op, err)
			}
			next = append(next, discovered...)
		}
		queue = next
	}
	return graph, nil
}

// expandGraph добавляет прямые ребра вершины и возвращает только новые вершины для очереди BFS.
// Уже загруженные пакеты берутся из packages; повторные ребра не добавляются.
func (a *Analyzer) expandGraph(graph *domain.Graph, id domain.PackageID,
	packages map[string]domain.PackageVersion) ([]domain.PackageID, error) {
	var discovered []domain.PackageID
	for _, dep := range graph.Nodes[id].Dependencies {
		pkg, loaded := packages[dep.Name]
		if !loaded {
			var err error
			pkg, err = a.PackageDependencies(dep.Name)
			if err != nil {
				return nil, fmt.Errorf("dependency %s of %s: %w", dep.Name, id.Name, err)
			}
			packages[dep.Name] = pkg
			packages[pkg.Name] = pkg
		}
		target := domain.PackageID{Name: pkg.Name, Version: pkg.Version}
		if !slices.Contains(graph.Edges[id], target) {
			graph.Edges[id] = append(graph.Edges[id], target)
		}
		if _, found := graph.Nodes[target]; !found {
			graph.Nodes[target] = pkg
			discovered = append(discovered, target)
		}
	}
	return discovered, nil
}

// selectVersion выбирает максимальную неотозванную версию одного пакета без prerelease.
// При равном приоритете возвращается первый подходящий элемент вместе с его зависимостями.
// Некорректный номер неотозванной версии или отсутствие подходящих версий дает ошибку.
func (a *Analyzer) selectVersion(versions []domain.PackageVersion) (domain.PackageVersion, error) {
	const op = "analyzer.Analyzer.selectVersion"
	var best domain.PackageVersion
	found := false
	for _, ver := range versions {
		if ver.Yanked {
			continue
		}
		parsed, err := parseVersion(ver.Version)
		if err != nil {
			return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
		}
		if len(parsed.prerelease) != noIdentifiers {
			continue
		}
		if !found {
			best, found = ver, true
			continue
		}
		order, err := CompareVersions(ver.Version, best.Version)
		if err != nil {
			return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
		}
		if order == versionGreater {
			best = ver
		}
	}
	if !found {
		return domain.PackageVersion{}, fmt.Errorf("%s: no non-yanked release versions", op)
	}
	return best, nil
}
