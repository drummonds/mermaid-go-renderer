package mermaid

import (
	"os"
	"strings"
	"testing"
)

func loadTestdata(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/c4/" + name)
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	return string(data)
}

func TestC4Container_01_Title(t *testing.T) {
	input := loadTestdata(t, "container_01_title.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	if g.Kind != DiagramC4 {
		t.Fatalf("kind = %s, want c4", g.Kind)
	}
	if g.C4Title != "Model Bank - Container Diagram" {
		t.Fatalf("title = %q", g.C4Title)
	}
	if len(g.C4Elements) != 0 {
		t.Fatalf("elements = %d, want 0", len(g.C4Elements))
	}

	// Render and verify we get title text, not boxes
	svg, err := Render(input)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	t.Logf("SVG output:\n%s", svg)
	if !strings.Contains(svg, "Model Bank") {
		t.Fatalf("SVG should contain title text")
	}
}

func TestC4Container_02_Person(t *testing.T) {
	input := loadTestdata(t, "container_02_person.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	if g.C4Title != "Model Bank - Container Diagram" {
		t.Fatalf("title = %q", g.C4Title)
	}
	if len(g.C4Elements) != 1 {
		t.Fatalf("elements = %d, want 1", len(g.C4Elements))
	}
	e := g.C4Elements[0]
	if e.Kind != C4Person {
		t.Fatalf("kind = %s, want person", e.Kind)
	}
	if e.Alias != "customer" {
		t.Fatalf("alias = %q", e.Alias)
	}
	if e.Label != "Bank Customer" {
		t.Fatalf("label = %q", e.Label)
	}
}

func TestC4Container_03_Ext(t *testing.T) {
	input := loadTestdata(t, "container_03_ext.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	if len(g.C4Elements) != 2 {
		t.Fatalf("elements = %d, want 2", len(g.C4Elements))
	}
	fps := g.C4Elements[1]
	if fps.Kind != C4SystemExt {
		t.Fatalf("fps kind = %s, want system_ext", fps.Kind)
	}
	if fps.Alias != "fps" {
		t.Fatalf("fps alias = %q", fps.Alias)
	}
	if fps.Description != "FPS simulator" {
		t.Fatalf("fps desc = %q", fps.Description)
	}
}

func TestC4Container_04_BoundaryOne(t *testing.T) {
	input := loadTestdata(t, "container_04_boundary_one.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	if len(g.C4Boundaries) != 1 {
		t.Fatalf("boundaries = %d, want 1", len(g.C4Boundaries))
	}
	b := g.C4Boundaries[0]
	if b.Alias != "bank" || b.Label != "Model Bank" {
		t.Fatalf("boundary = %+v", b)
	}
	// 3 elements: customer (top-level), web (in boundary), fps (top-level)
	if len(g.C4Elements) != 3 {
		t.Fatalf("elements = %d, want 3", len(g.C4Elements))
	}
	web := g.C4Elements[1]
	if web.Alias != "web" || web.Boundary != "bank" {
		t.Fatalf("web = %+v", web)
	}
	if web.Technology != "Go WASM + Bulma" {
		t.Fatalf("web tech = %q", web.Technology)
	}
	if web.Description != "Browser-based dashboard" {
		t.Fatalf("web desc = %q", web.Description)
	}
	// fps should be top-level (outside boundary)
	fps := g.C4Elements[2]
	if fps.Boundary != "" {
		t.Fatalf("fps should be top-level, boundary = %q", fps.Boundary)
	}
}

func TestC4Container_05_Containers(t *testing.T) {
	input := loadTestdata(t, "container_05_containers.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	// customer + web + server + sim + payments + fps = 6
	if len(g.C4Elements) != 6 {
		t.Fatalf("elements = %d, want 6", len(g.C4Elements))
	}
	// All 4 containers should be inside "bank" boundary
	for _, alias := range []string{"web", "server", "sim", "payments"} {
		found := false
		for _, e := range g.C4Elements {
			if e.Alias == alias {
				found = true
				if e.Boundary != "bank" {
					t.Fatalf("%s boundary = %q, want bank", alias, e.Boundary)
				}
			}
		}
		if !found {
			t.Fatalf("element %q not found", alias)
		}
	}
}

func TestC4Container_06_Db(t *testing.T) {
	input := loadTestdata(t, "container_06_db.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	// customer + web + server + sim + payments + ledger + fps = 7
	if len(g.C4Elements) != 7 {
		t.Fatalf("elements = %d, want 7", len(g.C4Elements))
	}
	var ledger C4Element
	for _, e := range g.C4Elements {
		if e.Alias == "ledger" {
			ledger = e
		}
	}
	if ledger.Kind != C4ContainerDb {
		t.Fatalf("ledger kind = %s", ledger.Kind)
	}
	if ledger.Technology != "go-luca" {
		t.Fatalf("ledger tech = %q", ledger.Technology)
	}
	if ledger.Boundary != "bank" {
		t.Fatalf("ledger boundary = %q", ledger.Boundary)
	}
}

func TestC4Container_07_Rels(t *testing.T) {
	input := loadTestdata(t, "container_07_rels.mmd")
	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	g := parsed.Graph
	if len(g.C4Rels) != 7 {
		t.Fatalf("rels = %d, want 7", len(g.C4Rels))
	}
	// Spot-check first rel
	r0 := g.C4Rels[0]
	if r0.From != "customer" || r0.To != "web" || r0.Label != "Uses" || r0.Technology != "HTTPS" {
		t.Fatalf("rel[0] = %+v", r0)
	}
	// Check a rel without technology
	r4 := g.C4Rels[4]
	if r4.From != "sim" || r4.To != "ledger" || r4.Label != "Records transactions" {
		t.Fatalf("rel[4] = %+v", r4)
	}
}

// TestC4Container_07_Render verifies the full pipeline produces SVG
func TestC4Container_07_Render(t *testing.T) {
	input := loadTestdata(t, "container_07_rels.mmd")
	svg, err := Render(input)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(svg) < 100 {
		t.Fatalf("SVG too short: %d bytes", len(svg))
	}
}
