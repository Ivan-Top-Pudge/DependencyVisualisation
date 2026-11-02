package main

import (
	"flag"
	"fmt"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
)

func main() {
	configPath := flag.String("config", "tests/example.ini", "Путь к файлу конфигурации")

	flag.Parse()

	cfg := config.MustLoadConfig(*configPath)
	fmt.Printf("Package: %s\nRepository: %s\nTestMode: %t\nASCIITree: %t\nMaxDepth: %d\n",
		cfg.Package, cfg.Repository, cfg.TestMode, cfg.ASCIITree, cfg.MaxDepth)
}
