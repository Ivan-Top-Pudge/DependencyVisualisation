// Package cargo загружает версии пакетов и их прямые зависимости из HTTP(S)-индекса Cargo.
package cargo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

const (
	singleCharNameLength = 1
	doubleCharNameLength = 2
	tripleCharNameLength = 3
)

var indexPackageName = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)

// Client получает и разбирает метаданные пакетов из указанного индекса Cargo.
type Client struct {
	client  *http.Client
	repoURL string
}

// NewClient создает клиент для базового HTTP(S)-адреса индекса Cargo.
// Корректность адреса проверяется при вызове LoadVersions.
func NewClient(url string) *Client {
	return &Client{
		client:  http.DefaultClient,
		repoURL: url,
	}
}

// LoadVersions возвращает все версии пакета из индекса вместе с прямыми зависимостями.
// Отозванные версии сохраняются; Name зависимости содержит настоящее имя пакета.
// Запрос и чтение ответа ограничены тайм-аутом в пять секунд.
// Ошибки URL, сети, HTTP-статуса и разбора возвращаются вызывающему коду.
// Пустой ответ дает пустой список версий без ошибки.
func (c *Client) LoadVersions(pkgName string) ([]domain.PackageVersion, error) {
	const op = "cargo.Client.LoadVersions"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	packageURL, err := c.buildURL(pkgName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, packageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: unexpected status code %d", op, resp.StatusCode)
	}

	versions, err := parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return versions, nil
}

// buildURL строит HTTP(S)-адрес записи индекса Cargo для указанного пакета.
// Имя приводится к нижнему регистру; существующий путь репозитория сохраняется.
func (c *Client) buildURL(pkgName string) (string, error) {
	const op = "cargo.Client.buildURL"
	if !indexPackageName.MatchString(pkgName) {
		return "", fmt.Errorf("%s: invalid package name %q", op, pkgName)
	}
	baseURL, err := url.Parse(c.repoURL)
	if err != nil {
		return "", fmt.Errorf("%s: invalid repository URL: %w", op, err)
	}
	if (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Hostname() == "" {
		return "", fmt.Errorf("%s: repository must be an HTTP(S) URL with a host", op)
	}

	name := strings.ToLower(pkgName)
	nameRunes := []rune(name)
	var parts []string
	switch len(nameRunes) {
	case singleCharNameLength:
		parts = []string{"1"}
	case doubleCharNameLength:
		parts = []string{"2"}
	case tripleCharNameLength:
		parts = []string{"3", string(nameRunes[:singleCharNameLength])}
	default:
		parts = []string{
			string(nameRunes[:doubleCharNameLength]),
			string(nameRunes[doubleCharNameLength : doubleCharNameLength*2]),
		}
	}
	parts = append(parts, name)
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return baseURL.JoinPath(parts...).String(), nil
}
