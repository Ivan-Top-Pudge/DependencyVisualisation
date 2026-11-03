package cargo

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestLoadVersions проверяет HTTP-запрос, разбор JSONL и значения всех возвращаемых полей.
func TestLoadVersions(t *testing.T) {
	data, err := os.ReadFile("../../../tests/cargo_versions.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	server := newMetadataServer(t, http.StatusOK, data)
	got, err := NewClient(server.URL).LoadVersions("demo")
	if err != nil {
		t.Fatalf("LoadVersions() failed: %v", err)
	}
	if want := expectedPackageVersions(); !reflect.DeepEqual(got, want) {
		t.Errorf("LoadVersions() = %+v, want %+v", got, want)
	}
}

// TestLoadVersionsHTTPError проверяет возврат ошибки статуса до разбора тела ответа.
func TestLoadVersionsHTTPError(t *testing.T) {
	server := newMetadataServer(t, http.StatusNotFound, []byte("not JSON"))
	versions, err := NewClient(server.URL).LoadVersions("demo")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("LoadVersions() error = %v, want HTTP 404", err)
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		t.Fatalf("LoadVersions() parsed the body of an HTTP error: %v", err)
	}
	if versions != nil {
		t.Errorf("LoadVersions() returned versions after an HTTP error: %+v", versions)
	}
}

// TestLoadVersionsInvalidJSON проверяет сохранение ошибки разбора успешного HTTP-ответа.
func TestLoadVersionsInvalidJSON(t *testing.T) {
	server := newMetadataServer(t, http.StatusOK, []byte("not JSON"))
	versions, err := NewClient(server.URL).LoadVersions("demo")
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("LoadVersions() error = %v, want a JSON syntax error", err)
	}
	if versions != nil {
		t.Errorf("LoadVersions() returned versions after a parse error: %+v", versions)
	}
}

// TestLoadVersionsConnectionError проверяет запрос к недоступному локальному серверу.
func TestLoadVersionsConnectionError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	client := NewClient(server.URL)
	server.Close()
	versions, err := client.LoadVersions("demo")
	var networkErr net.Error
	if !errors.As(err, &networkErr) {
		t.Fatalf("LoadVersions() error = %v, want a network error", err)
	}
	if versions != nil {
		t.Errorf("LoadVersions() returned versions after a network error: %+v", versions)
	}
}

// TestLoadVersionsTimeout проверяет тайм-аут чтения тела ответа и сохранение ошибки сети.
func TestLoadVersionsTimeout(t *testing.T) {
	const testTimeout = 100 * time.Millisecond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	client := NewClient(server.URL)
	client.client = server.Client()
	client.client.Timeout = testTimeout
	versions, err := client.LoadVersions("demo")
	var networkErr net.Error
	if !errors.As(err, &networkErr) || !networkErr.Timeout() {
		t.Fatalf("LoadVersions() error = %v, want a timeout", err)
	}
	if versions != nil {
		t.Errorf("LoadVersions() returned versions after a timeout: %+v", versions)
	}
}

// newMetadataServer создает сервер, проверяющий метод и путь запроса метаданных demo.
func newMetadataServer(t *testing.T, status int, body []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/de/mo/demo" {
			t.Errorf("request path = %q, want /de/mo/demo", r.URL.Path)
		}
		w.WriteHeader(status)
		if _, err := w.Write(body); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
