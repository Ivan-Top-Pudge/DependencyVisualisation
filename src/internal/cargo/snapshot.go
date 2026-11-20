package cargo

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// WalkSnapshot читает tar.gz и передает версии каждого пакета обработчику.
// Файлы индекса могут лежать в корне архива или в одном каталоге снимка.
func WalkSnapshot(input io.Reader, visit func([]domain.PackageVersion) error) error {
	stream, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer stream.Close()
	archive := tar.NewReader(stream)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			_, err = io.Copy(io.Discard, stream)
			return err
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg || !isIndexEntry(header.Name) {
			continue
		}
		if err := visitEntry(archive, header.Name, visit); err != nil {
			return fmt.Errorf("%s: %w", header.Name, err)
		}
	}
}

// isIndexEntry проверяет, соответствует ли путь записи имени пакета в индексе.
func isIndexEntry(entry string) bool {
	entry = strings.TrimPrefix(entry, "./")
	name := path.Base(entry)
	if !indexPackageName.MatchString(name) {
		return false
	}
	expected := indexPath(name)
	wrapper, relative, _ := strings.Cut(entry, "/")
	return entry == expected || (wrapper != "" && !strings.HasPrefix(wrapper, ".") && relative == expected)
}

// visitEntry проверяет соответствие имен JSONL имени файла и передает версии обработчику.
func visitEntry(r io.Reader, entry string, visit func([]domain.PackageVersion) error) error {
	versions, err := parse(r)
	if err != nil {
		return err
	}
	for _, pkg := range versions {
		if !strings.EqualFold(pkg.Name, path.Base(entry)) {
			return fmt.Errorf("unexpected package name %q", pkg.Name)
		}
	}
	return visit(versions)
}
