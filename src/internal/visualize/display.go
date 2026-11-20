package visualize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Ivan-Top-Pudge/DependencyVisualisation/src/internal/domain"
)

// Show сохраняет D2 и SVG в отдельном временном каталоге и открывает SVG в системном просмотрщике.
// При ошибке рендеринга или открытия сообщение содержит путь к доступному файлу.
func Show(graph domain.Graph) (string, error) {
	dir, err := os.MkdirTemp("", "visualiser-")
	if err != nil {
		return "", err
	}
	source := filepath.Join(dir, "graph.d2")
	image := filepath.Join(dir, "graph.svg")
	if err := os.WriteFile(source, []byte(D2(graph)), 0600); err != nil {
		return "", err
	}
	if output, err := exec.Command("d2", source, image).CombinedOutput(); err != nil {
		return "", fmt.Errorf("D2 (%s): %w %s", source, err, strings.TrimSpace(string(output)))
	}
	if err := openImage(image); err != nil {
		return "", fmt.Errorf("просмотр SVG (%s): %w", image, err)
	}
	return image, nil
}

// openImage передает путь к готовому SVG системному приложению без командной оболочки.
func openImage(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "Safari", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
