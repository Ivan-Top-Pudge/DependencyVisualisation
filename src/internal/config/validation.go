package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const minDepth = 0

var (
	cargoPackageName = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
	testPackageName  = regexp.MustCompile(`^[A-Z]+$`)
)

// validateConfig проверяет имя пакета, источник репозитория и допустимую глубину анализа.
func validateConfig(cfg Config) error {
	if !validatePackageName(cfg.Package, cfg.TestMode) {
		return fmt.Errorf("неверное значение package_name: %q", cfg.Package)
	}
	if err := validateRepository(cfg.Repository, cfg.TestMode); err != nil {
		return err
	}
	if cfg.MaxDepth < minDepth {
		return fmt.Errorf("значение max_depth должно быть не меньше %d", minDepth)
	}

	return nil
}

// validatePackageName проверяет формат имени пакета с учетом тестового режима.
func validatePackageName(name string, testMode bool) bool {
	if testMode {
		return testPackageName.MatchString(name)
	}

	return cargoPackageName.MatchString(name)
}

// validateRepository проверяет URL или путь к файлу в зависимости от режима.
func validateRepository(repository string, testMode bool) error {
	if strings.TrimSpace(repository) == "" {
		return fmt.Errorf("значение repository не должно быть пустым")
	}
	if testMode {
		return validateRepositoryFile(repository)
	}

	return validateRepositoryURL(repository)
}

// validateRepositoryFile проверяет существование и доступность обычного файла.
func validateRepositoryFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("не удалось найти файл repository %q: %v", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("repository должен указывать на обычный файл: %q", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл repository %q: %v", path, err)
	}
	defer file.Close()

	return nil
}

// validateRepositoryURL проверяет формат HTTP(S)-адреса и наличие имени сервера.
func validateRepositoryURL(repository string) error {
	parsedURL, err := url.ParseRequestURI(repository)
	if err != nil {
		return fmt.Errorf("неверный URL repository: %v", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL repository должен использовать http или https")
	}
	if parsedURL.Hostname() == "" {
		return fmt.Errorf("URL repository должен содержать имя сервера")
	}

	return nil
}
