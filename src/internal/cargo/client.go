// Package cargo читает записи sparse-индекса Cargo и полные снимки в формате tar.gz.
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

// NewClient принимает базовый URL для LoadVersions или URL полного снимка для OpenIndex.
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

	return baseURL.JoinPath(indexPath(pkgName)).String(), nil
}

// indexPath возвращает относительный путь файла пакета по правилам индекса Cargo.
func indexPath(pkgName string) string {
	name := strings.ToLower(pkgName)
	chars := []rune(name)
	switch len(chars) {
	case singleCharNameLength:
		return "1/" + name
	case doubleCharNameLength:
		return "2/" + name
	case tripleCharNameLength:
		return "3/" + string(chars[:singleCharNameLength]) + "/" + name
	default:
		first := string(chars[:doubleCharNameLength])
		second := string(chars[doubleCharNameLength : doubleCharNameLength*2])
		return first + "/" + second + "/" + name
	}
}

// OpenIndex открывает HTTP-поток архива полного индекса.
// Вызывающий код должен закрыть поток после чтения.
func (c *Client) OpenIndex(ctx context.Context) (io.ReadCloser, error) {
	const op = "cargo.Client.OpenIndex"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.repoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		if err = resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return nil, fmt.Errorf("%s: unexpected status: %d", op, resp.StatusCode)
	}

	return resp.Body, nil
}
