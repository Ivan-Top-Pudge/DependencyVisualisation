package main

import (
	"cmp"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/cargo"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/testrepo"
)

// main запускает анализ пакета Cargo по INI-настройкам и выводит результат.
func main() {
	configPath := flag.String("config", "tests/example.ini", "Путь к файлу конфигурации")

	flag.Parse()

	cfg := config.MustLoadConfig(*configPath)

	var source analyzer.PackageSource = cargo.NewClient(cfg.Repository)
	if cfg.TestMode {
		var err error
		source, err = testrepo.Load(cfg.Repository)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения тестового репозитория: %v\n", err)
			os.Exit(1)
		}
	}
	graph, err := analyzer.New(source).BuildGraph(cfg.Package, cfg.MaxDepth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при анализе зависимостей пакета: %v\n", err)
		os.Exit(1)
	}

	printGraph(graph)
}

// printGraph выводит список смежности графа, упорядочив пакеты по имени и версии.
func printGraph(graph domain.Graph) {
	const noDependencies = 0
	fmt.Printf("Пакет: %s %s\nГраф зависимостей:\n", graph.Root.Name, graph.Root.Version)
	var ids []domain.PackageID
	for id := range graph.Nodes {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b domain.PackageID) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Version, b.Version))
	})
	for _, id := range ids {
		var dependencies []string
		for _, target := range graph.Edges[id] {
			dependencies = append(dependencies, target.Name+" "+target.Version)
		}
		if len(dependencies) == noDependencies {
			dependencies = append(dependencies, "-")
		}
		fmt.Printf("  %s %s: %s\n", id.Name, id.Version, strings.Join(dependencies, ", "))
	}
}
