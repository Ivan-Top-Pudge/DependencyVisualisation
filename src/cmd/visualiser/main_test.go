package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/visualize"
)

const testRepository = "../../../tests/graph.json"

// TestLocalVisualization проверяет направление ребер, циклы, общие вершины и глубину в обоих форматах.
func TestLocalVisualization(t *testing.T) {
	cfg := config.Config{Package: "A", Repository: testRepository, TestMode: true, ASCIITree: true, MaxDepth: 3}
	var output strings.Builder
	err := run(cfg, &output, func(graph domain.Graph) (string, error) {
		const nodes = "n0: \"A 1.0.0\"\nn1: \"B 1.0.0\"\nn2: \"C 1.0.0\"\nn3: \"D 1.0.0\"\nn4: \"E 1.0.0\"\n"
		const edges = "n0 -> n1\nn0 -> n2\nn1 -> n3\nn2 -> n3\nn2 -> n0\nn3 -> n3\nn3 -> n4\n"
		if got := visualize.D2(graph); got != nodes+"n0.style.double-border: true\n"+edges {
			t.Fatalf("D2: %s", got)
		}
		return "graph.svg", nil
	})
	const tree = "A 1.0.0\n|-- B 1.0.0\n|   `-- D 1.0.0\n|       |-- D 1.0.0 (*)\n|       `-- E 1.0.0\n" +
		"`-- C 1.0.0\n    |-- D 1.0.0 (*)\n    `-- A 1.0.0 (*)\n"
	if want := tree + "SVG: graph.svg\n"; output.String() != want || err != nil {
		t.Fatalf("вывод %q, ошибка %v", output.String(), err)
	}
}

// TestCargoVisualization проверяет выбор корня из sparse-индекса, глубину 0 и отключенное дерево.
func TestCargoVisualization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/de/mo/demo" {
			t.Errorf("неожиданный запрос %s", r.URL.Path)
		}
		io.WriteString(w, `{"name":"demo","vers":"1.0.0","deps":[{"name":"leaf"}]}`)
	}))
	defer server.Close()
	cfg := config.Config{Package: "demo", Repository: server.URL, MaxDepth: 0}
	var output strings.Builder
	err := run(cfg, &output, func(graph domain.Graph) (string, error) {
		if got := visualize.D2(graph); got != "n0: \"demo 1.0.0\"\nn0.style.double-border: true\n" {
			t.Fatalf("D2: %s", got)
		}
		return "graph.svg", nil
	})
	if err != nil || output.String() != "SVG: graph.svg\n" {
		t.Fatalf("вывод %q, ошибка %v", output.String(), err)
	}
}

// TestVisualizationErrors проверяет ошибки источника и просмотрщика.
func TestVisualizationErrors(t *testing.T) {
	cfg := config.Config{Package: "UNKNOWN", Repository: testRepository, TestMode: true}
	var output strings.Builder
	if err := run(cfg, &output, nil); err == nil || output.String() != "" {
		t.Fatalf("неизвестный пакет: вывод %q, ошибка %v", output.String(), err)
	}
	cfg.Package = "Z"
	viewerErr := errors.New("viewer failed")
	show := func(domain.Graph) (string, error) { return "", viewerErr }
	if err := run(cfg, &output, show); !errors.Is(err, viewerErr) || output.String() != "" {
		t.Fatalf("ошибка показа: вывод %q, ошибка %v", output.String(), err)
	}
}
