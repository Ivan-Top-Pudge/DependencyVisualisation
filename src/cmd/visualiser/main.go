package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/analyzer"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// main запускает анализ пакета Cargo по INI-настройкам и выводит результат.
func main() {
	configPath := flag.String("config", "tests/example.ini", "Путь к файлу конфигурации")

	flag.Parse()

	cfg := config.MustLoadConfig(*configPath)

	analyzer := analyzer.New(cfg.Repository)

	vers, err := analyzer.PackageDependencies(cfg.Package)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при анализе зависимостей пакета: %v\n", err)
		os.Exit(1)
	}

	printPackageDeps(vers)
}

// printPackageDeps выводит имя, версию и все прямые зависимости пакета.
// Для пустого списка зависимостей выводится отдельное сообщение.
func printPackageDeps(pkg domain.PackageVersion) {
	const emptyDependencyCount = 0
	fmt.Printf("Пакет: %s %s\n", pkg.Name, pkg.Version)
	if len(pkg.Dependencies) > emptyDependencyCount {
		fmt.Println("Прямые зависимости:")
		for _, dep := range pkg.Dependencies {
			fmt.Printf("  %s %s\n", dep.Name, dep.Requirement)
		}
	} else {
		fmt.Println("Прямых зависимостей нет")
	}
}
