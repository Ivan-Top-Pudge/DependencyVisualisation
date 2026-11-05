package analyzer_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/testrepo"
)

// recordingSource записывает порядок загрузок для проверки BFS и отсутствия повторных запросов.
type recordingSource struct {
	analyzer.PackageSource
	calls []string
}

// LoadVersions записывает запрос и передает его тестовому репозиторию.
func (s *recordingSource) LoadVersions(name string) ([]domain.PackageVersion, error) {
	s.calls = append(s.calls, name)
	return s.PackageSource.LoadVersions(name)
}

// TestBuildGraph проверяет глубину, BFS, общие зависимости, циклы, петли и повторные ребра.
func TestBuildGraph(t *testing.T) {
	source, err := testrepo.Load("../../../tests/graph.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		root  string
		depth int
		graph string
		calls string
	}{
		{"Z", 5, "Z:", "Z"},
		{"A", 0, "A:", "A"},
		{"A", 1, "A:B,C;B:;C:", "A,B,C"},
		{"A", 2, "A:B,C;B:D;C:D,A;D:", "A,B,C,D"},
		{"A", 10, "A:B,C;B:D;C:D,A;D:D,E;E:", "A,B,C,D,E"},
		{"B", 1, "B:D;D:", "B,D"},
	}
	for _, tt := range tests {
		t.Run(tt.graph, func(t *testing.T) {
			recorded := &recordingSource{PackageSource: source}
			graph, err := analyzer.New(recorded).BuildGraph(tt.root, tt.depth)
			if err != nil {
				t.Fatal(err)
			}
			if got := graphSummary(graph); got != tt.graph || strings.Join(recorded.calls, ",") != tt.calls {
				t.Errorf("graph = %s, loads = %v; want %s, %s", got, recorded.calls, tt.graph, tt.calls)
			}
			if graph.Root != (domain.PackageID{Name: tt.root, Version: "1.0.0"}) {
				t.Errorf("unexpected root: %+v", graph.Root)
			}
		})
	}
}

// TestBuildGraphErrors проверяет недопустимую глубину и отсутствие корневого пакета.
func TestBuildGraphErrors(t *testing.T) {
	source, err := testrepo.Load("../../../tests/graph.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		root  string
		depth int
	}{{"A", -1}, {"MISSING", 1}} {
		if _, err := analyzer.New(source).BuildGraph(tt.root, tt.depth); err == nil {
			t.Errorf("BuildGraph(%q, %d) accepted invalid input", tt.root, tt.depth)
		}
	}
}

// graphSummary возвращает упорядоченные списки смежности, включая вершины без исходящих ребер.
func graphSummary(graph domain.Graph) string {
	var rows []string
	for id := range graph.Nodes {
		var targets []string
		for _, target := range graph.Edges[id] {
			targets = append(targets, target.Name)
		}
		rows = append(rows, id.Name+":"+strings.Join(targets, ","))
	}
	slices.Sort(rows)
	return strings.Join(rows, ";")
}
