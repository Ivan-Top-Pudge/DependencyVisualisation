// Команда visualiser показывает граф зависимостей Cargo-пакета или вершины тестового репозитория.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/cargo"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/testrepo"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/visualize"
)

// main запускает визуализацию по настройкам указанного INI-файла.
func main() {
	configPath := flag.String("config", "tests/example.ini", "Путь к файлу конфигурации")
	flag.Parse()
	if err := run(config.MustLoadConfig(*configPath), os.Stdout, visualize.Show); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка визуализации: %v\n", err)
		os.Exit(1)
	}
}

// run строит обычный граф зависимостей, выводит запрошенное дерево и показывает диаграмму.
func run(cfg config.Config, output io.Writer, show func(domain.Graph) (string, error)) error {
	var source analyzer.PackageSource = cargo.NewClient(cfg.Repository)
	if cfg.TestMode {
		repo, err := testrepo.Load(cfg.Repository)
		if err != nil {
			return err
		}
		source = repo
	}
	graph, err := analyzer.New(source).BuildGraph(cfg.Package, cfg.MaxDepth)
	if err != nil {
		return err
	}
	if cfg.ASCIITree {
		if err := visualize.WriteTree(output, graph); err != nil {
			return err
		}
	}
	path, err := show(graph)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, "SVG:", path)
	return err
}
