// Package domain содержит структуры, представляющие пакеты и их зависимости.
package domain

// PackageVersion представляет собой версию пакета с его зависимостями.
type PackageVersion struct {
	Name         string       `json:"name"`
	Version      string       `json:"vers"`
	Yanked       bool         `json:"yanked"`
	Dependencies []Dependency `json:"deps"`
}

// Dependency описывает прямую зависимость и требование к ее версии.
// После разбора индекса Name содержит настоящее имя пакета.
type Dependency struct {
	Name        string `json:"name"`
	Requirement string `json:"req"`
	PackageName string `json:"package,omitempty"` // Настоящее имя при переименовании зависимости.
}
