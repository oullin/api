package staticgen

import (
	"fmt"
	baseHttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/oullin/metal/kernel"
)

type Generator struct {
	OutputDir string
}

func NewGenerator(outputDir string) Generator {
	return Generator{OutputDir: outputDir}
}

func (g Generator) Generate(routes []kernel.StaticRouteDefinition) ([]string, error) {
	if strings.TrimSpace(g.OutputDir) == "" {
		return nil, fmt.Errorf("output directory must be provided")
	}

	if err := os.MkdirAll(g.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	generated := make([]string, 0, len(routes))

	for _, route := range routes {
		if strings.TrimSpace(route.Path) == "" {
			return nil, fmt.Errorf("static route path cannot be empty")
		}

		if route.Maker == nil {
			return nil, fmt.Errorf("static route %s is missing a handler", route.Path)
		}

		resource := route.Maker(route.File)
		if resource == nil {
			return nil, fmt.Errorf("static route %s returned no handler", route.Path)
		}

		request := httptest.NewRequest(baseHttp.MethodGet, route.Path, nil)
		recorder := httptest.NewRecorder()

		if apiErr := resource.Handle(recorder, request); apiErr != nil {
			return nil, fmt.Errorf("static route %s failed: %s", route.Path, apiErr.Message)
		}

		if recorder.Code != baseHttp.StatusOK {
			return nil, fmt.Errorf("static route %s returned status %d", route.Path, recorder.Code)
		}

		body := recorder.Body.Bytes()
		if len(body) == 0 {
			return nil, fmt.Errorf("static route %s returned an empty body", route.Path)
		}

		fileName := strings.Trim(route.Path, "/")
		if fileName == "" {
			fileName = "index"
		}

		filePath := filepath.Join(g.OutputDir, fileName+".json")

		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return nil, fmt.Errorf("creating directory for %s: %w", filePath, err)
		}

		if err := os.WriteFile(filePath, body, 0o644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", filePath, err)
		}

		generated = append(generated, filePath)
	}

	return generated, nil
}
