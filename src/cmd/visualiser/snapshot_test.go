package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/config"
)

// TestCargoSnapshot проверяет HTTP, чтение снимка, политику версий и обход через CLI.
func TestCargoSnapshot(t *testing.T) {
	files := map[string]string{
		"index/config.json":     "служебный файл",
		"index/.github/1/a":     "служебный файл",
		"index/se/rd/serde":     `{"name":"serde","vers":"1.0.0"}`,
		"index/de/mo/demo":      demoVersions,
		"index/3/a/app":         `{"name":"app","vers":"0.1.0","deps":[{"name":"Demo"}]}`,
		"index/1/y":             `{"name":"y","vers":"1.0.0","yanked":true,"deps":[{"name":"serde"}]}`,
		"index/1/p":             `{"name":"p","vers":"1.0.0-rc","deps":[{"name":"serde"}]}`,
		"index/1/x":             `{"name":"x","vers":"1.0.0","deps":[{"name":"serde","registry":"https://other"}]}`,
		"index/not/index/serde": "служебный файл",
	}
	url := snapshotServer(t, http.StatusOK, makeSnapshot(t, files))
	cfg := config.Config{Package: "Serde", Repository: url, MaxDepth: 2}
	var output strings.Builder
	if err := run(cfg, &output); err != nil {
		t.Fatal(err)
	}
	const want = "Обратные зависимости пакета serde 1.0.0:\n  Demo 1.10.0\n  app 0.1.0\n"
	if output.String() != want {
		t.Fatalf("получено %q, ожидалось %q", output.String(), want)
	}
}

const demoVersions = `{"name":"Demo","vers":"1.10.0","deps":[{"name":"alias","package":"serde"},{"name":"serde"}]}
{"name":"Demo","vers":"2.0.0","yanked":true}
{"name":"Demo","vers":"3.0.0-beta"}
{"name":"Demo","vers":"1.9.0"}`

// TestCargoSnapshotErrors проверяет сообщения об ошибках и отсутствие частичного вывода.
func TestCargoSnapshotErrors(t *testing.T) {
	const valid = `{"name":"a","vers":"1.0.0"}`
	corrupt := makeSnapshot(t, map[string]string{"1/a": valid})
	corrupt[len(corrupt)-1] ^= 0xff
	tests := []struct {
		name   string
		status int
		data   []byte
		want   string
	}{
		{"http", http.StatusNotFound, nil, "404"},
		{"gzip", http.StatusOK, []byte("invalid input"), "gzip"},
		{"tar", http.StatusOK, compressData(t, []byte("bad tar")), "tar"},
		{"checksum", http.StatusOK, corrupt, "checksum"},
		{"json", http.StatusOK, makeSnapshot(t, map[string]string{"1/a": "{"}), "1/a"},
		{"name", http.StatusOK, makeSnapshot(t, map[string]string{"1/b": valid}), "unexpected package name"},
		{"version", http.StatusOK, makeSnapshot(t, map[string]string{"1/a": `{"name":"a","vers":"v1"}`}), "semantic"},
		{"no_release", http.StatusOK, makeSnapshot(t, map[string]string{"1/a": demoYanked}), "no stable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{Package: "a", Repository: snapshotServer(t, tt.status, tt.data)}
			var output strings.Builder
			err := run(cfg, &output)
			if err == nil || !strings.Contains(err.Error(), tt.want) || output.String() != "" {
				t.Errorf("получено %q, %v, ожидалась ошибка %q", output.String(), err, tt.want)
			}
		})
	}
}

const demoYanked = `{"name":"a","vers":"1.0.0","yanked":true}`

// snapshotServer публикует один снимок по HTTP для проверки обычного режима.
func snapshotServer(t *testing.T, status int, data []byte) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/snapshot" || r.Method != http.MethodGet {
			t.Errorf("неожиданный запрос %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(status)
		if _, err := w.Write(data); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL + "/snapshot"
}

// makeSnapshot упаковывает файлы миниатюрного индекса в tar.gz в памяти.
func makeSnapshot(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	for name, data := range files {
		header := tar.Header{Name: name, Mode: 0644, Size: int64(len(data))}
		if err := w.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return compressData(t, buf.Bytes())
}

// compressData сжимает тестовый поток для проверки формата и целостности архива.
func compressData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
