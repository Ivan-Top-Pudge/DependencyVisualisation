// Package visualize строит D2-диаграмму и ASCII-дерево по одному графу зависимостей.
package visualize

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// D2 описывает каждую вершину и каждое направленное ребро, включая петли.
// Числовые идентификаторы отделяют синтаксис D2 от имен и версий пакетов.
func D2(graph domain.Graph) string {
	nodes := slices.SortedFunc(maps.Keys(graph.Nodes), func(a, b domain.PackageID) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Version, b.Version))
	})
	ids := make(map[domain.PackageID]int, len(nodes))
	var source strings.Builder
	for i, node := range nodes {
		ids[node] = i
		quoted := strconv.Quote(node.Name + " " + node.Version)
		fmt.Fprintf(&source, "n%d: %s\n", i, strings.ReplaceAll(quoted, "$", `\$`))
	}
	fmt.Fprintf(&source, "n%d.style.double-border: true\n", ids[graph.Root])
	for _, node := range nodes {
		for _, dependency := range graph.Edges[node] {
			fmt.Fprintf(&source, "n%d -> n%d\n", ids[node], ids[dependency])
		}
	}
	return source.String()
}
