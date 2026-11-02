package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfigINI = `[visualizer]
package_name = serde
repository = https://example.org/index
test_mode = false
ascii_tree = false
max_depth = 3
`

// configSuccessCase описывает корректный INI и ожидаемые значения всех параметров.
type configSuccessCase struct {
	name     string
	contents string
	want     Config
}

// configErrorCase описывает ошибочный INI и ожидаемый фрагмент сообщения об ошибке.
type configErrorCase struct {
	name      string
	contents  string
	wantError string
}

// TestLoadConfig проверяет значения всех полей в обычном и тестовом режимах.
func TestLoadConfig(t *testing.T) {
	repository := writeConfigTestFile(t, "repository.txt", "A\n")
	for _, test := range validConfigCases(repository) {
		t.Run(test.name, func(t *testing.T) {
			path := writeConfigTestFile(t, "config.ini", test.contents)
			cfg, err := loadConfig(path)
			if err != nil {
				t.Fatalf("не удалось загрузить корректную конфигурацию: %v", err)
			}
			if cfg != test.want {
				t.Fatalf("получена конфигурация %+v, ожидалась %+v", cfg, test.want)
			}
		})
	}
}

// TestLoadConfigErrors проверяет недопустимые значения и отсутствие каждого параметра.
func TestLoadConfigErrors(t *testing.T) {
	repository := writeConfigTestFile(t, "repository.txt", "A\n")
	cases := append(invalidConfigCases(repository), missingConfigCases()...)
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := writeConfigTestFile(t, "config.ini", test.contents)
			_, err := loadConfig(path)
			if err == nil {
				t.Fatalf("ожидалась ошибка с фрагментом %q", test.wantError)
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("ошибка %q не содержит %q", err, test.wantError)
			}
		})
	}
}

// TestLoadConfigFileErrors проверяет отсутствующий файл и каталог вместо INI-файла.
func TestLoadConfigFileErrors(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"missing_file", filepath.Join(t.TempDir(), "missing.ini")},
		{"directory", t.TempDir()},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadConfig(test.path)
			if err == nil {
				t.Fatal("ожидалась ошибка чтения файла")
			}
			if !strings.Contains(err.Error(), "не удалось прочитать файл") {
				t.Fatalf("получена неожиданная ошибка: %v", err)
			}
		})
	}
}

// validConfigCases задает ожидаемые результаты для обоих режимов и булевых значений.
func validConfigCases(repository string) []configSuccessCase {
	local := localConfigINI(repository)
	zeroDepth := changeConfigValue(validConfigINI, "max_depth", "0")
	zeroDepth = changeConfigValue(zeroDepth, "ascii_tree", "true")
	return []configSuccessCase{
		{
			name: "normal_mode", contents: validConfigINI,
			want: Config{
				Package: "serde", Repository: "https://example.org/index",
				TestMode: false, ASCIITree: false, MaxDepth: 3,
			},
		},
		{
			name: "zero_depth_and_ascii", contents: zeroDepth,
			want: Config{
				Package: "serde", Repository: "https://example.org/index",
				TestMode: false, ASCIITree: true, MaxDepth: 0,
			},
		},
		{
			name: "test_mode", contents: local,
			want: Config{
				Package: "A", Repository: repository,
				TestMode: true, ASCIITree: false, MaxDepth: 3,
			},
		},
		{
			name: "test_mode_and_ascii", contents: changeConfigValue(local, "ascii_tree", "true"),
			want: Config{
				Package: "A", Repository: repository,
				TestMode: true, ASCIITree: true, MaxDepth: 3,
			},
		},
	}
}

// invalidConfigCases задает недопустимые значения параметров и ошибки структуры INI.
func invalidConfigCases(repository string) []configErrorCase {
	local := localConfigINI(repository)
	missing := filepath.Join(filepath.Dir(repository), "missing.txt")
	return []configErrorCase{
		{"empty_package", changeConfigValue(validConfigINI, "package_name", ""), "package_name"},
		{"blank_package", changeConfigValue(validConfigINI, "package_name", `"   "`), "package_name"},
		{"invalid_package", changeConfigValue(validConfigINI, "package_name", "bad/name"), "package_name"},
		{"lowercase_test_package", changeConfigValue(local, "package_name", "serde"), "package_name"},
		{"empty_repository", changeConfigValue(validConfigINI, "repository", ""), "repository"},
		{"blank_repository", changeConfigValue(validConfigINI, "repository", `"   "`), "repository"},
		{"invalid_url", changeConfigValue(validConfigINI, "repository", "not-a-url"), "repository"},
		{"invalid_url_escape", changeConfigValue(validConfigINI, "repository", "https://host/%zz"), "repository"},
		{"missing_hostname", changeConfigValue(validConfigINI, "repository", "https:///index"), "repository"},
		{"unsupported_scheme", changeConfigValue(validConfigINI, "repository", "ftp://host/index"), "repository"},
		{"missing_repository_file", changeConfigValue(local, "repository", missing), "repository"},
		{"repository_directory", changeConfigValue(local, "repository", filepath.Dir(repository)), "repository"},
		{"empty_test_mode", changeConfigValue(validConfigINI, "test_mode", ""), "test_mode"},
		{"invalid_test_mode", changeConfigValue(validConfigINI, "test_mode", "banana"), "test_mode"},
		{"empty_ascii_tree", changeConfigValue(validConfigINI, "ascii_tree", ""), "ascii_tree"},
		{"invalid_ascii_tree", changeConfigValue(validConfigINI, "ascii_tree", "banana"), "ascii_tree"},
		{"empty_depth", changeConfigValue(validConfigINI, "max_depth", ""), "max_depth"},
		{"negative_depth", changeConfigValue(validConfigINI, "max_depth", "-1"), "max_depth"},
		{"fractional_depth", changeConfigValue(validConfigINI, "max_depth", "1.5"), "max_depth"},
		{"nonnumeric_depth", changeConfigValue(validConfigINI, "max_depth", "abc"), "max_depth"},
		{"overflow_depth", changeConfigValue(validConfigINI, "max_depth", "999999999999999999999"), "max_depth"},
		{"missing_section", "[other]\n", "visualizer"},
		{"malformed_ini", "[visualizer\n", "не удалось прочитать файл"},
	}
}

// missingConfigCases задает конфигурации без одного из пяти обязательных ключей.
func missingConfigCases() []configErrorCase {
	var cases []configErrorCase
	for _, key := range []string{"package_name", "repository", "test_mode", "ascii_tree", "max_depth"} {
		cases = append(cases, configErrorCase{
			name: "missing_" + key, contents: removeConfigKey(validConfigINI, key), wantError: key,
		})
	}
	return cases
}

// writeConfigTestFile создает временный файл для отдельного тестового случая.
func writeConfigTestFile(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}
	return path
}

// localConfigINI возвращает корректный INI тестового режима с заданным файлом репозитория.
func localConfigINI(repository string) string {
	return fmt.Sprintf(`[visualizer]
package_name = A
repository = %s
test_mode = true
ascii_tree = false
max_depth = 3
`, repository)
}

// changeConfigValue заменяет значение одного ключа в тестовом INI.
func changeConfigValue(contents, key, value string) string {
	lines := strings.Split(contents, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key+" =") {
			lines[i] = key + " = " + value
		}
	}
	return strings.Join(lines, "\n")
}

// removeConfigKey удаляет один ключ из тестового INI.
func removeConfigKey(contents, key string) string {
	lines := strings.Split(contents, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key+" =") {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}
