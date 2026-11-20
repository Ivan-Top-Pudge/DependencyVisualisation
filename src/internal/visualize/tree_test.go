package visualize

import (
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// TestShortestTreePath проверяет ссылку на вершину, которая раскрывается позже, но ближе к корню.
func TestShortestTreePath(t *testing.T) {
	a, b := domain.PackageID{Name: "A", Version: "1"}, domain.PackageID{Name: "B", Version: "1"}
	c, d := domain.PackageID{Name: "C", Version: "1"}, domain.PackageID{Name: "D", Version: "1"}
	graph := domain.Graph{Root: a, Edges: map[domain.PackageID][]domain.PackageID{a: {b, c}, b: {c}, c: {d}}}
	var tree strings.Builder
	const want = "A 1\n|-- B 1\n|   `-- C 1 (*)\n`-- C 1\n    `-- D 1\n"
	if err := WriteTree(&tree, graph); err != nil || tree.String() != want {
		t.Fatalf("дерево %q, ошибка %v", tree.String(), err)
	}
}
