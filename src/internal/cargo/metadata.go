package cargo

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// parse читает версии пакета из JSONL непосредственно в модели domain.
// Для переименованных зависимостей Name заменяется настоящим именем из PackageName.
func parse(r io.Reader) ([]domain.PackageVersion, error) {
	const op = "cargo.parse"
	var packages []domain.PackageVersion
	decoder := json.NewDecoder(r)
	for {
		var pkg domain.PackageVersion
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		for i := range pkg.Dependencies {
			dependency := &pkg.Dependencies[i]
			if dependency.PackageName != "" {
				dependency.Name = dependency.PackageName
			}
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}
