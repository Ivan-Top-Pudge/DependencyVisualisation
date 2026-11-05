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

// PackageID различает вершины графа по имени и выбранной версии пакета.
type PackageID struct {
	Name    string
	Version string
}

// Graph хранит корень, пакеты и направленные ребра к их зависимостям.
// Зависимости в Nodes содержат исходные метаданные; границы обхода отражены в Edges.
type Graph struct {
	Root  PackageID
	Nodes map[PackageID]PackageVersion
	Edges map[PackageID][]PackageID
}
