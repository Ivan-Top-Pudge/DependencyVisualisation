// Package analyzer содержит операции анализа метаданных пакетов Cargo.
package analyzer

import (
	"cmp"
	"fmt"
	"regexp"
	"strings"
)

const (
	versionLess     = -1
	versionEqual    = 0
	versionGreater  = 1
	coreParts       = 3
	majorGroup      = 1
	minorGroup      = 2
	patchGroup      = 3
	prereleaseGroup = 4
	noIdentifiers   = 0
)

const (
	numericIdentifierPattern    = `(?:0|[1-9][0-9]*)`
	prereleaseIdentifierPattern = `(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)`
	buildIdentifierPattern      = `[0-9A-Za-z-]+`
	semanticVersionPattern      = `^(` + numericIdentifierPattern + `)\.` +
		`(` + numericIdentifierPattern + `)\.(` + numericIdentifierPattern + `)` +
		`(?:-(` + prereleaseIdentifierPattern + `(?:\.` + prereleaseIdentifierPattern + `)*))?` +
		`(?:\+` + buildIdentifierPattern + `(?:\.` + buildIdentifierPattern + `)*)?$`
)

var (
	semanticVersionRE   = regexp.MustCompile(semanticVersionPattern)
	numericIdentifierRE = regexp.MustCompile(`^` + numericIdentifierPattern + `$`)
)

// semanticVersion хранит проверенные компоненты версии без метаданных сборки.
type semanticVersion struct {
	core       [coreParts]string
	prerelease []string
}

// CompareVersions сравнивает две версии по правилам [SemVer 2.0.0].
// Возвращает -1, если left меньше right, 0 при равном приоритете и 1, если left больше.
// Предварительная версия ниже соответствующего релиза; метаданные после + не влияют на порядок.
// Числовые компоненты могут иметь произвольную длину: преобразование в int не используется.
// Если любая строка не соответствует формату SemVer, возвращается ошибка.
//
// [SemVer 2.0.0]: https://semver.org/
func CompareVersions(left, right string) (int, error) {
	const op = "analyzer.CompareVersions"
	a, err := parseVersion(left)
	if err != nil {
		return versionEqual, fmt.Errorf("%s: left version: %w", op, err)
	}
	b, err := parseVersion(right)
	if err != nil {
		return versionEqual, fmt.Errorf("%s: right version: %w", op, err)
	}
	for i := range a.core {
		if order := compareNumericIdentifier(a.core[i], b.core[i]); order != versionEqual {
			return order, nil
		}
	}
	return comparePrerelease(a.prerelease, b.prerelease), nil
}

// parseVersion проверяет формат SemVer и выделяет основные и предварительные компоненты.
func parseVersion(value string) (semanticVersion, error) {
	matches := semanticVersionRE.FindStringSubmatch(value)
	if matches == nil {
		return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
	}
	version := semanticVersion{
		core: [coreParts]string{matches[majorGroup], matches[minorGroup], matches[patchGroup]},
	}
	if matches[prereleaseGroup] != "" {
		version.prerelease = strings.Split(matches[prereleaseGroup], ".")
	}
	return version, nil
}

// compareNumericIdentifier сравнивает десятичные числа без ведущих нулей и переполнения.
func compareNumericIdentifier(left, right string) int {
	if order := cmp.Compare(len(left), len(right)); order != versionEqual {
		return order
	}
	return strings.Compare(left, right)
}

// comparePrerelease сравнивает предварительные части; пустая часть обозначает обычный релиз.
func comparePrerelease(left, right []string) int {
	if len(left) == noIdentifiers && len(right) == noIdentifiers {
		return versionEqual
	}
	if len(left) == noIdentifiers {
		return versionGreater
	}
	if len(right) == noIdentifiers {
		return versionLess
	}
	for i := range min(len(left), len(right)) {
		if order := comparePrereleaseIdentifier(left[i], right[i]); order != versionEqual {
			return order
		}
	}
	return cmp.Compare(len(left), len(right))
}

// comparePrereleaseIdentifier сравнивает числа численно, остальные идентификаторы - по ASCII.
func comparePrereleaseIdentifier(left, right string) int {
	leftNumeric := numericIdentifierRE.MatchString(left)
	rightNumeric := numericIdentifierRE.MatchString(right)
	if leftNumeric && rightNumeric {
		return compareNumericIdentifier(left, right)
	}
	if leftNumeric {
		return versionLess
	}
	if rightNumeric {
		return versionGreater
	}
	return strings.Compare(left, right)
}
