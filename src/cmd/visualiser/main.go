// Команда visualiser выводит обратные зависимости пакета из заданного в INI репозитория.
package main

import (
	"bufio"
	"cmp"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/cargo"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/reverse"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/testrepo"
)

// main запускает поиск обратных зависимостей по INI-настройкам.
func main() {
	configPath := flag.String("config", "tests/example.ini", "Путь к файлу конфигурации")
	flag.Parse()
	if err := run(config.MustLoadConfig(*configPath), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при поиске обратных зависимостей: %v\n", err)
		os.Exit(1)
	}
}

// run строит индекс входящих ребер и обходит его от заданного пакета до max_depth.
func run(cfg config.Config, output io.Writer) error {
	source := reverse.New()
	if err := visitRepository(cfg, source.AddVersions); err != nil {
		return err
	}
	graph, err := analyzer.New(source).BuildGraph(cfg.Package, cfg.MaxDepth)
	if err != nil {
		return err
	}
	return printDependents(output, graph)
}

// visitRepository перебирает пакеты локального JSON-графа или HTTP-снимка индекса Cargo.
func visitRepository(cfg config.Config, visit func([]domain.PackageVersion) error) error {
	if cfg.TestMode {
		repo, err := testrepo.Load(cfg.Repository)
		if err != nil {
			return err
		}
		return repo.VisitVersions(visit)
	}
	stream, err := cargo.NewClient(cfg.Repository).OpenIndex(context.Background())
	if err != nil {
		return err
	}
	defer stream.Close()
	return cargo.WalkSnapshot(stream, visit)
}

// printDependents выводит найденные пакеты в порядке имен и версий, исключая корень.
func printDependents(output io.Writer, graph domain.Graph) error {
	const noDependents = 0
	var ids []domain.PackageID
	for id := range graph.Nodes {
		if id != graph.Root {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, func(a, b domain.PackageID) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Version, b.Version))
	})
	w := bufio.NewWriter(output)
	fmt.Fprintf(w, "Обратные зависимости пакета %s %s:\n", graph.Root.Name, graph.Root.Version)
	if len(ids) == noDependents {
		fmt.Fprintln(w, "  Не найдены.")
	}
	for _, id := range ids {
		fmt.Fprintf(w, "  %s %s\n", id.Name, id.Version)
	}
	return w.Flush()
}
