package mermaid

import "strings"

func parseC4(input string) (ParseOutput, error) {
	lines, err := preprocessInput(input)
	if err != nil {
		return ParseOutput{}, err
	}
	graph := newGraph(DiagramC4)
	graph.Source = input
	graph.C4ShapesPerRow = 4
	graph.C4BoundariesPerRow = 2

	var boundaryStack []string // stack of boundary aliases

	for idx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if idx == 0 && isHeaderLineForKind(line, DiagramC4) {
			continue
		}
		if line == "" {
			continue
		}

		l := lower(line)

		// title
		if strings.HasPrefix(l, "title ") {
			graph.C4Title = strings.TrimSpace(line[6:])
			continue
		}

		// closing brace — pop boundary
		if line == "}" {
			if len(boundaryStack) > 0 {
				boundaryStack = boundaryStack[:len(boundaryStack)-1]
			}
			continue
		}

		// UpdateLayoutConfig
		if strings.HasPrefix(l, "updatelayoutconfig(") {
			args := parseC4Args(line)
			for _, arg := range args {
				a := strings.TrimSpace(arg)
				if strings.HasPrefix(a, "$c4ShapeInRow") {
					if v := c4ConfigValue(a); v > 0 {
						graph.C4ShapesPerRow = v
					}
				}
				if strings.HasPrefix(a, "$c4BoundaryInRow") {
					if v := c4ConfigValue(a); v > 0 {
						graph.C4BoundariesPerRow = v
					}
				}
			}
			continue
		}

		// Skip style directives
		if strings.HasPrefix(l, "updateelementstyle(") ||
			strings.HasPrefix(l, "updaterelstyle(") {
			continue
		}

		// Boundaries
		if bAlias, bLabel, ok := parseC4Boundary(line); ok {
			parent := ""
			if len(boundaryStack) > 0 {
				parent = boundaryStack[len(boundaryStack)-1]
			}
			graph.C4Boundaries = append(graph.C4Boundaries, C4Boundary{
				Alias:  bAlias,
				Label:  bLabel,
				Parent: parent,
			})
			boundaryStack = append(boundaryStack, bAlias)
			continue
		}

		// Relationships
		if rel, ok := parseC4Rel(line); ok {
			graph.C4Rels = append(graph.C4Rels, rel)
			continue
		}

		// Elements
		if elem, ok := parseC4Element(line); ok {
			if len(boundaryStack) > 0 {
				elem.Boundary = boundaryStack[len(boundaryStack)-1]
			}
			graph.C4Elements = append(graph.C4Elements, elem)
			continue
		}
	}

	return ParseOutput{Graph: graph}, nil
}

// parseC4Args extracts arguments from a C4 function call like Person(alias, "Label", "Desc").
// Finds the opening '(' and splits on commas respecting quotes.
func parseC4Args(line string) []string {
	start := strings.Index(line, "(")
	if start < 0 {
		return nil
	}
	// Find matching closing paren
	depth := 0
	end := -1
	for i := start; i < len(line); i++ {
		switch line[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		end = len(line) // no closing paren, take rest
	}
	content := line[start+1 : end]

	var args []string
	var current strings.Builder
	var quote byte
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if quote != 0 {
			if ch == quote {
				quote = 0
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}
		if ch == ',' {
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		args = append(args, s)
	}
	return args
}

func parseC4Element(line string) (C4Element, bool) {
	l := lower(line)
	keyword := ""
	var kind C4ElementKind

	// Order matters: check longer prefixes first
	elementMap := []struct {
		prefix string
		kind   C4ElementKind
	}{
		{"person_ext(", C4PersonExt},
		{"person(", C4Person},
		{"systemdb_ext(", C4SystemDbExt},
		{"systemqueue_ext(", C4SystemQueueExt},
		{"system_ext(", C4SystemExt},
		{"systemdb(", C4SystemDb},
		{"systemqueue(", C4SystemQueue},
		{"system(", C4System},
		{"containerdb_ext(", C4ContainerDbExt},
		{"containerqueue_ext(", C4ContainerQueueExt},
		{"container_ext(", C4ContainerExt},
		{"containerdb(", C4ContainerDb},
		{"containerqueue(", C4ContainerQueue},
		{"container(", C4Container},
		{"componentdb_ext(", C4ComponentDbExt},
		{"componentqueue_ext(", C4ComponentQueueExt},
		{"component_ext(", C4ComponentExt},
		{"componentdb(", C4ComponentDb},
		{"componentqueue(", C4ComponentQueue},
		{"component(", C4Component},
	}

	for _, em := range elementMap {
		if strings.HasPrefix(l, em.prefix) {
			keyword = em.prefix
			kind = em.kind
			break
		}
	}
	if keyword == "" {
		return C4Element{}, false
	}

	args := parseC4Args(line)
	if len(args) < 1 {
		return C4Element{}, false
	}

	elem := C4Element{
		Alias: strings.TrimSpace(args[0]),
		Kind:  kind,
	}
	if len(args) >= 2 {
		elem.Label = args[1]
	}
	// For containers/components, 3rd arg is technology, 4th is description
	// For person/system, 3rd arg is description
	hasTech := kind == C4Container || kind == C4ContainerDb || kind == C4ContainerQueue ||
		kind == C4ContainerExt || kind == C4ContainerDbExt || kind == C4ContainerQueueExt ||
		kind == C4Component || kind == C4ComponentDb || kind == C4ComponentQueue ||
		kind == C4ComponentExt || kind == C4ComponentDbExt || kind == C4ComponentQueueExt

	if hasTech {
		if len(args) >= 3 {
			elem.Technology = args[2]
		}
		if len(args) >= 4 {
			elem.Description = args[3]
		}
	} else {
		if len(args) >= 3 {
			elem.Description = args[2]
		}
	}

	if elem.Label == "" {
		elem.Label = elem.Alias
	}
	return elem, true
}

func parseC4Boundary(line string) (alias, label string, ok bool) {
	l := lower(line)
	isBoundary := false

	prefixes := []string{
		"system_boundary(",
		"enterprise_boundary(",
		"boundary(",
		"container_boundary(",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(l, p) {
			isBoundary = true
			break
		}
	}
	if !isBoundary {
		return "", "", false
	}

	// Strip trailing { if present
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimRight(trimmed, " \t{")

	args := parseC4Args(trimmed)
	if len(args) < 1 {
		return "", "", false
	}
	alias = strings.TrimSpace(args[0])
	if len(args) >= 2 {
		label = args[1]
	}
	if label == "" {
		label = alias
	}
	return alias, label, true
}

func parseC4Rel(line string) (C4Rel, bool) {
	l := lower(line)

	type relPrefix struct {
		prefix string
		dir    string
		bidir  bool
	}
	prefixes := []relPrefix{
		{"birel(", "", true},
		{"rel_up(", "up", false},
		{"rel_u(", "up", false},
		{"rel_down(", "down", false},
		{"rel_d(", "down", false},
		{"rel_left(", "left", false},
		{"rel_l(", "left", false},
		{"rel_right(", "right", false},
		{"rel_r(", "right", false},
		{"rel_back(", "back", false},
		{"rel(", "", false},
	}

	var matched *relPrefix
	for i := range prefixes {
		if strings.HasPrefix(l, prefixes[i].prefix) {
			matched = &prefixes[i]
			break
		}
	}
	if matched == nil {
		return C4Rel{}, false
	}

	args := parseC4Args(line)
	if len(args) < 2 {
		return C4Rel{}, false
	}

	rel := C4Rel{
		From:      strings.TrimSpace(args[0]),
		To:        strings.TrimSpace(args[1]),
		Direction: matched.dir,
		BiDir:     matched.bidir,
	}
	if len(args) >= 3 {
		rel.Label = args[2]
	}
	if len(args) >= 4 {
		rel.Technology = args[3]
	}
	return rel, true
}

func c4ConfigValue(arg string) int {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		return 0
	}
	v, ok := parseFloat(strings.TrimSpace(parts[1]))
	if !ok || v < 1 {
		return 0
	}
	return int(v)
}
