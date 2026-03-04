package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mermaid "github.com/bvolpato/mermaid-go-renderer"
)

func main() {
	testdataDir := "testdata"
	outputFile := "docs/demos.html"

	var files []string
	err := filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".mmd") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walking testdata: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(files)

	var sections strings.Builder
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "reading %s: %v\n", f, err)
			continue
		}
		svg, err := mermaid.Render(string(src))
		if err != nil {
			fmt.Fprintf(os.Stderr, "rendering %s: %v\n", f, err)
			continue
		}
		sections.WriteString(fmt.Sprintf(`    <section class="section">
      <h2 class="title is-4">%s</h2>
      <div class="columns">
        <div class="column is-half">
          <pre><code>%s</code></pre>
        </div>
        <div class="column is-half">
          %s
        </div>
      </div>
    </section>
`, html.EscapeString(f), html.EscapeString(string(src)), svg))
	}

	page := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>mermaid-go-renderer demos</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  <style>
    pre { white-space: pre-wrap; word-break: break-word; }
    svg { max-width: 100%%; height: auto; }
  </style>
</head>
<body>
  <div class="container">
    <h1 class="title is-2 mt-5">mermaid-go-renderer demos</h1>
%s  </div>
</body>
</html>
`, sections.String())

	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "creating output dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outputFile, []byte(page), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "writing %s: %v\n", outputFile, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d diagrams)\n", outputFile, len(files))
}
