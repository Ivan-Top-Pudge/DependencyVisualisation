// Package config загружает и проверяет INI-настройки визуализатора зависимостей.
package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// Config содержит параметры секции [visualizer] INI-файла.
type Config struct {
	Package    string `ini:"package_name"` // Имя Cargo-пакета; в тестовом режиме - заглавные латинские буквы.
	Repository string `ini:"repository"`   // HTTP(S)-URL или путь к читаемому файлу тестового репозитория.
	TestMode   bool   `ini:"test_mode"`    // Использовать файл тестового репозитория вместо URL.
	ASCIITree  bool   `ini:"ascii_tree"`   // Выводить зависимости в виде ASCII-дерева.
	MaxDepth   int    `ini:"max_depth"`    // Максимальная глубина анализа; глубина корневого пакета равна нулю.
}

// MustLoadConfig читает и проверяет конфигурацию из INI-файла.
// При ошибке чтения или проверки функция вызывает panic.
func MustLoadConfig(path string) Config {
	cfg, err := loadConfig(path)
	if err != nil {
		panic(err)
	}

	return cfg
}

// loadConfig читает секцию [visualizer] и проверяет значения параметров.
// Ошибки чтения и проверки возвращаются вызывающему коду.
func loadConfig(path string) (Config, error) {
	inidata, err := ini.Load(path)
	if err != nil {
		return Config{}, fmt.Errorf("не удалось прочитать файл: %v", err)
	}

	section, err := inidata.GetSection("visualizer")
	if err != nil {
		return Config{}, fmt.Errorf("не удалось прочитать раздел visualizer: %v", err)
	}

	cfg, err := readConfig(section)
	if err != nil {
		return Config{}, err
	}
	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// readConfig читает параметры секции, преобразуя булевы значения и глубину к типам Config.
// Допустимость имени пакета, источника и диапазона глубины проверяет validateConfig.
func readConfig(section *ini.Section) (Config, error) {
	testMode, err := section.Key("test_mode").Bool()
	if err != nil {
		return Config{}, fmt.Errorf("не удалось прочитать значение test_mode: %v", err)
	}

	ASCIITree, err := section.Key("ascii_tree").Bool()
	if err != nil {
		return Config{}, fmt.Errorf("не удалось прочитать значение ascii_tree: %v", err)
	}

	maxDepth, err := section.Key("max_depth").Int()
	if err != nil {
		return Config{}, fmt.Errorf("не удалось прочитать числовое значение max_depth: %v", err)
	}

	return Config{
		Package:    section.Key("package_name").String(),
		Repository: section.Key("repository").String(),
		TestMode:   testMode,
		ASCIITree:  ASCIITree,
		MaxDepth:   maxDepth,
	}, nil
}
