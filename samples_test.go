package mermaid

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestSamplesRender(t *testing.T) {
	entries, err := os.ReadDir("samples")
	if err != nil {
		t.Fatalf("read samples directory: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mmd") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		t.Fatal("expected at least one .mmd file in samples")
	}

	for _, name := range names {
		name := name
		t.Run(strings.TrimSuffix(name, ".mmd"), func(t *testing.T) {
			path := filepath.Join("samples", name)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read sample %s: %v", path, err)
			}

			svg, err := RenderWithOptions(
				string(content),
				DefaultRenderOptions().WithAllowApproximate(true),
			)
			if err != nil {
				t.Fatalf("render sample %s: %v", path, err)
			}
			if !strings.Contains(svg, "<svg") {
				t.Fatalf("sample %s did not produce SVG output", path)
			}

			width, height := detectSVGSize(svg)
			if width <= 0 || height <= 0 {
				width, height = 1600, 1200
			}
			img, err := rasterizeSVGToImage(svg, width, height)
			if err != nil {
				t.Fatalf("rasterize sample %s: %v", path, err)
			}
			if img == nil || img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
				t.Fatalf("sample %s produced invalid PNG image bounds", path)
			}
		})
	}
}
