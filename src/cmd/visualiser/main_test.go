package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
)

// TestReverseGraph проверяет вывод CLI на локальном графе при разной глубине.
func TestReverseGraph(t *testing.T) {
	tests := []struct {
		root  string
		depth int
		want  string
	}{
		{"Z", 3, ""}, {"E", 0, ""}, {"E", 1, "D"}, {"E", 2, "B C D"},
		{"E", 3, "A B C D"}, {"A", 30, "C"}, {"D", 1, "B C"}, {"B", 1, "A"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%d", tt.root, tt.depth), func(t *testing.T) {
			cfg := config.Config{Package: tt.root, Repository: "../../../tests/graph.json", TestMode: true, MaxDepth: tt.depth}
			var output strings.Builder
			if err := run(cfg, &output); err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("Обратные зависимости пакета %s 1.0.0:\n", tt.root)
			if tt.want == "" {
				want += "  Не найдены.\n"
			}
			for _, name := range strings.Fields(tt.want) {
				want += "  " + name + " 1.0.0\n"
			}
			if output.String() != want {
				t.Errorf("получено %q, ожидалось %q", output.String(), want)
			}
		})
	}
}

// TestUnknownRoot проверяет отсутствие частичного вывода при неизвестном пакете.
func TestUnknownRoot(t *testing.T) {
	cfg := config.Config{Repository: "../../../tests/graph.json", TestMode: true, Package: "UNKNOWN"}
	var output strings.Builder
	if err := run(cfg, &output); err == nil || output.String() != "" {
		t.Fatalf("неизвестный пакет: вывод %q, ошибка %v", output.String(), err)
	}
}
