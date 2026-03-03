package mermaid

import (
	"strings"
	"testing"
)

func TestRenderSimpleFlowchart(t *testing.T) {
	input := "flowchart LR\nA[Start] -->|go| B(End)"
	svg, err := Render(input)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Fatalf("expected SVG output, got: %s", svg)
	}
}

func TestRenderAllDiagramKinds(t *testing.T) {
	cases := []struct {
		name    string
		kind    DiagramKind
		diagram string
	}{
		{"flowchart", DiagramFlowchart, "flowchart TD\nA --> B --> C"},
		{"sequence", DiagramSequence, "sequenceDiagram\nparticipant Alice\nparticipant Bob\nAlice->>Bob: Hello"},
		{"class", DiagramClass, "classDiagram\nAnimal <|-- Duck\nDuck : +swim()"},
		{"state", DiagramState, "stateDiagram-v2\n[*] --> Active\nActive --> [*]"},
		{"er", DiagramER, "erDiagram\nCAR ||--o{ DRIVER : allows"},
		{"pie", DiagramPie, "pie showData\ntitle Pets\nDogs : 10\nCats : 5"},
		{"mindmap", DiagramMindmap, "mindmap\n  root((Mindmap))\n    child one\n    child two"},
		{"journey", DiagramJourney, "journey\ntitle Checkout Journey\nsection Happy\nBrowse: 5: Customer"},
		{"timeline", DiagramTimeline, "timeline\ntitle Product\n2024 : MVP\n2025 : GA"},
		{"gantt", DiagramGantt, "gantt\ntitle Roadmap\nsection Build\nCore Engine :done, core, 2026-01-01, 10d"},
		{"requirement", DiagramRequirement, "requirementDiagram\nrequirement req1 {\n id: 1\n text: fast renderer\n}"},
		{"gitgraph", DiagramGitGraph, "gitGraph\ncommit\nbranch feature\ncheckout feature\ncommit\ncheckout main\nmerge feature"},
		{"c4", DiagramC4, "C4Context\n    Person(customer, \"Bank Customer\", \"A customer\")\n    System(bank, \"Model Bank\", \"Core banking\")\n    System_Ext(fps, \"Faster Payments\", \"FPS mock\")\n    Rel(customer, bank, \"Uses\")\n    Rel(bank, fps, \"Sends payments\")"},
		{"sankey", DiagramSankey, "sankey-beta\nA,B,10"},
		{"quadrant", DiagramQuadrant, "quadrantChart\ntitle Priorities\nx-axis Low --> High\ny-axis Low --> High\nRisk: [0.2, 0.9]"},
		{"zenuml", DiagramZenUML, "zenuml\n@startuml\nAlice->Bob: ping"},
		{"block", DiagramBlock, "block-beta\ncolumns 2\nA B"},
		{"packet", DiagramPacket, "packet-beta\npacket test {\nfield a\n}"},
		{"kanban", DiagramKanban, "kanban\nTodo:\n  task one"},
		{"architecture", DiagramArchitecture, "architecture-beta\ngroup api(cloud)[API]"},
		{"radar", DiagramRadar, "radar-beta\ntitle Skills\naxis A, B, C\nseries You: [8,7,9]"},
		{"treemap", DiagramTreemap, "treemap\ntitle Market\nA: 10\nB: 20"},
		{"xychart", DiagramXYChart, "xychart-beta\ntitle Revenue\nx-axis [Q1, Q2, Q3]\ny-axis 0 --> 100\nbar [20, 50, 80]\nline [10, 40, 90]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseMermaid(tc.diagram)
			if err != nil {
				t.Fatalf("ParseMermaid() error = %v", err)
			}
			if parsed.Graph.Kind != tc.kind {
				t.Fatalf("kind mismatch: got %s, want %s", parsed.Graph.Kind, tc.kind)
			}
			layout := ComputeLayout(&parsed.Graph, ModernTheme(), DefaultLayoutConfig())
			if layout.Width <= 0 || layout.Height <= 0 {
				t.Fatalf("invalid layout size: %fx%f", layout.Width, layout.Height)
			}
			svg := RenderSVG(layout, ModernTheme(), DefaultLayoutConfig())
			if !strings.Contains(svg, "<svg") {
				t.Fatalf("expected SVG for kind %s", tc.kind)
			}
		})
	}
}

func TestRenderWithTiming(t *testing.T) {
	result, err := RenderWithTiming("flowchart LR\nA-->B-->C", DefaultRenderOptions())
	if err != nil {
		t.Fatalf("RenderWithTiming() error = %v", err)
	}
	if result.TotalUS() == 0 {
		t.Fatalf("expected non-zero timing, got %d", result.TotalUS())
	}
	if !strings.Contains(result.SVG, "<svg") {
		t.Fatalf("expected SVG output")
	}
}

func TestPreferredAspectRatio(t *testing.T) {
	opts := DefaultRenderOptions().WithPreferredAspectRatio(16.0 / 9.0)
	svg, err := RenderWithOptions("flowchart LR\nA-->B-->C-->D", opts)
	if err != nil {
		t.Fatalf("RenderWithOptions() error = %v", err)
	}
	width, okW := parseSVGAttr(svg, "width")
	height, okH := parseSVGAttr(svg, "height")
	if !okW || !okH || height == 0 {
		t.Fatalf("missing width/height in SVG")
	}
	ratio := width / height
	if ratio < 1.7 || ratio > 1.85 {
		t.Fatalf("expected near 16:9 ratio, got %f", ratio)
	}
}

func TestExtractMermaidBlocks(t *testing.T) {
	markdown := "text\n" +
		"``` mermaid\n" +
		"flowchart LR\n" +
		"  A --> B\n" +
		"```\n" +
		"other\n" +
		"~~~ mermaid\n" +
		"sequenceDiagram\n" +
		"  A->>B: hi\n" +
		"~~~\n"
	blocks := ExtractMermaidBlocks(markdown)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if !strings.Contains(blocks[0], "flowchart") || !strings.Contains(blocks[1], "sequenceDiagram") {
		t.Fatalf("unexpected block extraction: %#v", blocks)
	}
}

func TestParseC4SystemContext(t *testing.T) {
	input := `C4Context
    title Model Bank - System Context
    Person(customer, "Bank Customer", "A customer of the model bank")
    System(modelBank, "Model Bank", "Simulates core banking")
    System_Ext(fps, "Faster Payments", "UK FPS")
    Rel(customer, modelBank, "Views accounts")
    Rel(modelBank, fps, "Sends payments")`

	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("ParseMermaid() error = %v", err)
	}
	g := parsed.Graph
	if g.Kind != DiagramC4 {
		t.Fatalf("kind = %s, want c4", g.Kind)
	}
	if g.C4Title != "Model Bank - System Context" {
		t.Fatalf("title = %q", g.C4Title)
	}
	if len(g.C4Elements) != 3 {
		t.Fatalf("elements = %d, want 3", len(g.C4Elements))
	}
	if g.C4Elements[0].Kind != C4Person || g.C4Elements[0].Alias != "customer" {
		t.Fatalf("element[0] = %+v", g.C4Elements[0])
	}
	if g.C4Elements[1].Kind != C4System || g.C4Elements[1].Alias != "modelBank" {
		t.Fatalf("element[1] = %+v", g.C4Elements[1])
	}
	if g.C4Elements[2].Kind != C4SystemExt || g.C4Elements[2].Alias != "fps" {
		t.Fatalf("element[2] = %+v", g.C4Elements[2])
	}
	if len(g.C4Rels) != 2 {
		t.Fatalf("rels = %d, want 2", len(g.C4Rels))
	}
}

func TestParseC4Container(t *testing.T) {
	input := `C4Container
    title Container Diagram
    Person(customer, "Bank Customer")
    System_Boundary(bank, "Model Bank") {
        Container(web, "Web UI", "Go WASM", "Dashboard")
        ContainerDb(ledger, "Ledger", "go-luca", "Double-entry")
    }
    System_Ext(fps, "mock-fps", "FPS simulator")
    Rel(customer, web, "Uses", "HTTPS")`

	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("ParseMermaid() error = %v", err)
	}
	g := parsed.Graph
	if len(g.C4Elements) != 4 {
		t.Fatalf("elements = %d, want 4", len(g.C4Elements))
	}
	if len(g.C4Boundaries) != 1 {
		t.Fatalf("boundaries = %d, want 1", len(g.C4Boundaries))
	}
	if g.C4Boundaries[0].Alias != "bank" || g.C4Boundaries[0].Label != "Model Bank" {
		t.Fatalf("boundary = %+v", g.C4Boundaries[0])
	}
	// web and ledger should be inside the boundary
	if g.C4Elements[1].Boundary != "bank" || g.C4Elements[2].Boundary != "bank" {
		t.Fatalf("boundary assignment: web=%q ledger=%q", g.C4Elements[1].Boundary, g.C4Elements[2].Boundary)
	}
	// Container with technology
	if g.C4Elements[1].Technology != "Go WASM" {
		t.Fatalf("web technology = %q", g.C4Elements[1].Technology)
	}
	// ContainerDb
	if g.C4Elements[2].Kind != C4ContainerDb {
		t.Fatalf("ledger kind = %s", g.C4Elements[2].Kind)
	}
}

func TestC4LayoutProducesSVG(t *testing.T) {
	input := `C4Context
    Person(customer, "Bank Customer", "A customer")
    System(bank, "Model Bank", "Core banking")
    Rel(customer, bank, "Uses")`

	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("ParseMermaid() error = %v", err)
	}
	layout := ComputeLayout(&parsed.Graph, ModernTheme(), DefaultLayoutConfig())
	if layout.Width <= 0 || layout.Height <= 0 {
		t.Fatalf("invalid layout size: %fx%f", layout.Width, layout.Height)
	}
	svg := RenderSVG(layout, ModernTheme(), DefaultLayoutConfig())
	if !strings.Contains(svg, "<svg") {
		t.Fatalf("expected SVG output")
	}
	// Should contain C4 colors
	if !strings.Contains(svg, "#08427B") && !strings.Contains(svg, "#1168BD") {
		t.Fatalf("expected C4 colors in SVG")
	}
}

func TestC4BiRel(t *testing.T) {
	input := `C4Context
    System(a, "A")
    System(b, "B")
    BiRel(a, b, "Syncs")`

	parsed, err := ParseMermaid(input)
	if err != nil {
		t.Fatalf("ParseMermaid() error = %v", err)
	}
	if len(parsed.Graph.C4Rels) != 1 {
		t.Fatalf("rels = %d, want 1", len(parsed.Graph.C4Rels))
	}
	if !parsed.Graph.C4Rels[0].BiDir {
		t.Fatalf("expected BiDir=true")
	}
}

func parseSVGAttr(svg, attr string) (float64, bool) {
	marker := attr + "=\""
	start := strings.Index(svg, marker)
	if start < 0 {
		return 0, false
	}
	start += len(marker)
	end := strings.Index(svg[start:], "\"")
	if end < 0 {
		return 0, false
	}
	return parseFloat(svg[start : start+end])
}
