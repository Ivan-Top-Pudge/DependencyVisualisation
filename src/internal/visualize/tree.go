package visualize

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// parentsByDepth выбирает первого родителя на кратчайшем пути от корня для каждой вершины.
func parentsByDepth(graph domain.Graph) map[domain.PackageID]domain.PackageID {
	parents := map[domain.PackageID]domain.PackageID{graph.Root: graph.Root}
	queue := []domain.PackageID{graph.Root}
	for i := 0; i < len(queue); i++ {
		id := queue[i]
		for _, child := range graph.Edges[id] {
			if _, found := parents[child]; !found {
				parents[child] = id
				queue = append(queue, child)
			}
		}
	}
	return parents
}

// WriteTree печатает зависимости с ASCII-ветвями без рекурсии. Каждая вершина
// раскрывается один раз на минимальной глубине. Остальные ссылки получают метку (*).
func WriteTree(output io.Writer, graph domain.Graph) error {
	type branch struct {
		id    domain.PackageID
		width int
		next  int
	}
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, graph.Root.Name, graph.Root.Version)
	parents := parentsByDepth(graph)
	stack := []branch{{id: graph.Root}}
	var prefix bytes.Buffer
	const empty = 0
	for len(stack) > empty {
		last := len(stack) - 1
		parent := &stack[last]
		children := graph.Edges[parent.id]
		if parent.next == len(children) {
			stack = stack[:last]
			continue
		}
		child := children[parent.next]
		parent.next++
		arm, indent := "|-- ", "|   "
		if parent.next == len(children) {
			arm, indent = "`-- ", "    "
		}
		prefix.Truncate(parent.width)
		fmt.Fprintf(w, "%s%s%s %s", prefix.Bytes(), arm, child.Name, child.Version)
		if child == graph.Root || parents[child] != parent.id {
			fmt.Fprintln(w, " (*)")
			continue
		}
		fmt.Fprintln(w)
		prefix.WriteString(indent)
		stack = append(stack, branch{id: child, width: prefix.Len()})
	}
	return w.Flush()
}
