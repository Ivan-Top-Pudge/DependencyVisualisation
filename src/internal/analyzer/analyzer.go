package analyzer

import (
	"fmt"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/cargo"
	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// Analyzer объединяет операции анализа версий пакетов Cargo.
type Analyzer struct {
	client *cargo.Client
}

// New создает анализатор для базового HTTP(S)-адреса индекса Cargo.
// Адрес проверяется при первом вызове PackageDependencies.
func New(url string) *Analyzer {
	return &Analyzer{
		client: cargo.NewClient(url),
	}
}

// PackageDependencies загружает версии пакета и возвращает выбранную с прямыми зависимостями.
// Выбирается максимальная по SemVer неотозванная версия без prerelease.
// Ошибки загрузки, разбора и выбора версии возвращаются с контекстом операции.
func (a *Analyzer) PackageDependencies(name string) (domain.PackageVersion, error) {
	const op = "analyzer.Analyzer.PackageDependencies"
	versions, err := a.client.LoadVersions(name)
	if err != nil {
		return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
	}

	selectedVersion, err := a.selectVersion(versions)
	if err != nil {
		return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
	}

	return selectedVersion, nil
}

// selectVersion выбирает максимальную неотозванную версию одного пакета без prerelease.
// При равном приоритете возвращается первый подходящий элемент вместе с его зависимостями.
// Некорректный номер неотозванной версии или отсутствие подходящих версий дает ошибку.
func (a *Analyzer) selectVersion(versions []domain.PackageVersion) (domain.PackageVersion, error) {
	const op = "analyzer.Analyzer.selectVersion"
	var best domain.PackageVersion
	found := false
	for _, ver := range versions {
		if ver.Yanked {
			continue
		}
		parsed, err := parseVersion(ver.Version)
		if err != nil {
			return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
		}
		if len(parsed.prerelease) != noIdentifiers {
			continue
		}
		if !found {
			best, found = ver, true
			continue
		}
		order, err := CompareVersions(ver.Version, best.Version)
		if err != nil {
			return domain.PackageVersion{}, fmt.Errorf("%s: %w", op, err)
		}
		if order == versionGreater {
			best = ver
		}
	}
	if !found {
		return domain.PackageVersion{}, fmt.Errorf("%s: no non-yanked release versions", op)
	}
	return best, nil
}
