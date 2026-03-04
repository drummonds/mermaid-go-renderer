package mermaid

import "math"

// C4 color palette
const (
	c4PersonFill       = "#08427B"
	c4PersonExtFill    = "#999999"
	c4SystemFill       = "#1168BD"
	c4SystemExtFill    = "#999999"
	c4ContainerFill    = "#438DD5"
	c4ContainerExtFill = "#999999"
	c4ComponentFill    = "#85BBF0"
	c4ComponentExtFill = "#999999"
	c4TextWhite        = "#ffffff"
	c4TextDark         = "#000000"
	c4BoundaryStroke   = "#444444"
)

func c4ElementColors(kind C4ElementKind) (fill, text string) {
	switch kind {
	case C4Person:
		return c4PersonFill, c4TextWhite
	case C4PersonExt:
		return c4PersonExtFill, c4TextWhite
	case C4System, C4SystemDb, C4SystemQueue:
		return c4SystemFill, c4TextWhite
	case C4SystemExt, C4SystemDbExt, C4SystemQueueExt:
		return c4SystemExtFill, c4TextWhite
	case C4Container, C4ContainerDb, C4ContainerQueue:
		return c4ContainerFill, c4TextWhite
	case C4ContainerExt, C4ContainerDbExt, C4ContainerQueueExt:
		return c4ContainerExtFill, c4TextWhite
	case C4Component, C4ComponentDb, C4ComponentQueue:
		return c4ComponentFill, c4TextDark
	case C4ComponentExt, C4ComponentDbExt, C4ComponentQueueExt:
		return c4ComponentExtFill, c4TextWhite
	default:
		return c4SystemFill, c4TextWhite
	}
}

func isPersonKind(kind C4ElementKind) bool {
	return kind == C4Person || kind == C4PersonExt
}

func isDbKind(kind C4ElementKind) bool {
	switch kind {
	case C4SystemDb, C4SystemDbExt, C4ContainerDb, C4ContainerDbExt, C4ComponentDb, C4ComponentDbExt:
		return true
	}
	return false
}

type c4Positioned struct {
	elem C4Element
	x, y float64
	w, h float64
}

func layoutC4(graph *Graph, theme Theme, _ LayoutConfig) Layout {
	layout := Layout{Kind: graph.Kind}

	padding := 40.0
	elemW := 200.0
	elemH := 120.0
	elemSpacingX := 40.0
	elemSpacingY := 40.0
	boundaryPad := 30.0
	boundaryLabelH := 30.0
	shapesPerRow := graph.C4ShapesPerRow
	if shapesPerRow < 1 {
		shapesPerRow = 4
	}
	boundariesPerRow := graph.C4BoundariesPerRow
	if boundariesPerRow < 1 {
		boundariesPerRow = 2
	}

	// Index elements by boundary
	topLevel := []C4Element{}
	byBoundary := map[string][]C4Element{}
	for _, elem := range graph.C4Elements {
		if elem.Boundary == "" {
			topLevel = append(topLevel, elem)
		} else {
			byBoundary[elem.Boundary] = append(byBoundary[elem.Boundary], elem)
		}
	}

	// Track positioned elements by alias for relationship routing
	positioned := map[string]c4Positioned{}
	cursorY := padding

	// Title
	if graph.C4Title != "" {
		layout.Texts = append(layout.Texts, LayoutText{
			X:      padding,
			Y:      cursorY + 20,
			Value:  graph.C4Title,
			Anchor: "start",
			Size:   theme.FontSize + 4,
			Weight: "600",
			Color:  theme.PrimaryTextColor,
		})
		cursorY += 40
	}

	// Layout top-level elements in a row
	if len(topLevel) > 0 {
		rowW := layoutC4Row(topLevel, padding, cursorY, elemW, elemH, elemSpacingX, shapesPerRow, positioned)
		_ = rowW
		rows := (len(topLevel) + shapesPerRow - 1) / shapesPerRow
		cursorY += float64(rows)*(elemH+elemSpacingY) + elemSpacingY
	}

	// Layout boundaries
	for bIdx, boundary := range graph.C4Boundaries {
		children := byBoundary[boundary.Alias]
		if len(children) == 0 {
			continue
		}

		bx := padding
		// Arrange boundaries in grid
		col := bIdx % boundariesPerRow
		if col > 0 {
			// Compute width of previous boundaries in this row
			// Simple: stack boundaries horizontally
			bx = padding + float64(col)*(elemW*float64(shapesPerRow)+elemSpacingX*float64(shapesPerRow)+boundaryPad*2+elemSpacingX)
		}

		childRows := (len(children) + shapesPerRow - 1) / shapesPerRow
		innerW := float64(min(len(children), shapesPerRow))*(elemW+elemSpacingX) - elemSpacingX
		innerH := float64(childRows)*(elemH+elemSpacingY) - elemSpacingY
		bw := innerW + boundaryPad*2
		bh := innerH + boundaryPad*2 + boundaryLabelH

		// Draw boundary rect
		layout.Rects = append(layout.Rects, LayoutRect{
			X:           bx,
			Y:           cursorY,
			W:           bw,
			H:           bh,
			RX:          8,
			RY:          8,
			Fill:        "none",
			Stroke:      c4BoundaryStroke,
			StrokeWidth: 2,
			Dashed:      true,
		})
		// Boundary label
		layout.Texts = append(layout.Texts, LayoutText{
			X:      bx + boundaryPad,
			Y:      cursorY + 20,
			Value:  boundary.Label,
			Anchor: "start",
			Size:   theme.FontSize + 1,
			Weight: "600",
			Color:  c4BoundaryStroke,
		})

		// Layout children inside boundary
		layoutC4Row(children, bx+boundaryPad, cursorY+boundaryLabelH+boundaryPad, elemW, elemH, elemSpacingX, shapesPerRow, positioned)

		if bIdx%boundariesPerRow == boundariesPerRow-1 || bIdx == len(graph.C4Boundaries)-1 {
			cursorY += bh + elemSpacingY
		}
	}

	// Render element boxes
	for _, elem := range graph.C4Elements {
		pos, ok := positioned[elem.Alias]
		if !ok {
			continue
		}
		fill, textColor := c4ElementColors(elem.Kind)

		if isPersonKind(elem.Kind) {
			addC4Person(&layout, pos.x, pos.y, pos.w, pos.h, fill)
		} else if isDbKind(elem.Kind) {
			// Cylinder shape for database types
			addC4Cylinder(&layout, pos.x, pos.y, pos.w, pos.h, fill)
		} else {
			layout.Rects = append(layout.Rects, LayoutRect{
				X:           pos.x,
				Y:           pos.y,
				W:           pos.w,
				H:           pos.h,
				RX:          8,
				RY:          8,
				Fill:        fill,
				Stroke:      fill,
				StrokeWidth: 1,
			})
		}

		// Multi-line text: Label (bold), [Technology], Description
		// Person shapes have a head circle on top, so push text down
		lineY := pos.y + 30
		if isPersonKind(elem.Kind) {
			lineY = pos.y + 48
		}
		// Label
		layout.Texts = append(layout.Texts, LayoutText{
			X:      pos.x + pos.w/2,
			Y:      lineY,
			Value:  elem.Label,
			Anchor: "middle",
			Size:   theme.FontSize,
			Weight: "600",
			Color:  textColor,
		})
		lineY += theme.FontSize + 4
		// Technology
		if elem.Technology != "" {
			layout.Texts = append(layout.Texts, LayoutText{
				X:      pos.x + pos.w/2,
				Y:      lineY,
				Value:  "[" + elem.Technology + "]",
				Anchor: "middle",
				Size:   max(10, theme.FontSize-2),
				Color:  textColor,
			})
			lineY += theme.FontSize
		}
		// Description
		if elem.Description != "" {
			layout.Texts = append(layout.Texts, LayoutText{
				X:      pos.x + pos.w/2,
				Y:      lineY,
				Value:  elem.Description,
				Anchor: "middle",
				Size:   max(10, theme.FontSize-2),
				Color:  textColor,
			})
		}
	}

	// Collect boundary rects for crossing detection
	var boundaryRects []LayoutRect
	for _, r := range layout.Rects {
		if r.Dashed {
			boundaryRects = append(boundaryRects, r)
		}
	}

	// Optimize relationship connections to minimize crossings
	optimized := c4OptimizeConnections(graph.C4Rels, positioned, boundaryRects)

	for i, rel := range graph.C4Rels {
		pts := optimized[i]
		x1, y1, x2, y2 := pts[0], pts[1], pts[2], pts[3]
		if x1 == 0 && y1 == 0 && x2 == 0 && y2 == 0 {
			continue // from/to not found
		}

		layout.Lines = append(layout.Lines, LayoutLine{
			X1:          x1,
			Y1:          y1,
			X2:          x2,
			Y2:          y2,
			Stroke:      "#707070",
			StrokeWidth: 1.5,
			Dashed:      true,
			ArrowStart:  rel.BiDir,
			ArrowEnd:    true,
		})

		// Relationship label at midpoint
		midX := (x1 + x2) / 2
		midY := (y1 + y2) / 2
		if rel.Label != "" {
			layout.Texts = append(layout.Texts, LayoutText{
				X:      midX,
				Y:      midY - 6,
				Value:  rel.Label,
				Anchor: "middle",
				Size:   max(10, theme.FontSize-2),
				Color:  theme.PrimaryTextColor,
			})
		}
		if rel.Technology != "" {
			layout.Texts = append(layout.Texts, LayoutText{
				X:      midX,
				Y:      midY + 10,
				Value:  "[" + rel.Technology + "]",
				Anchor: "middle",
				Size:   max(9, theme.FontSize-3),
				Color:  theme.PrimaryTextColor,
			})
		}
	}

	// Compute total size — start with cursorY so title-only diagrams get height
	maxX := padding + measureTextWidth(graph.C4Title, true)
	maxY := cursorY
	for _, pos := range positioned {
		if pos.x+pos.w > maxX {
			maxX = pos.x + pos.w
		}
		if pos.y+pos.h > maxY {
			maxY = pos.y + pos.h
		}
	}
	for _, rect := range layout.Rects {
		if rect.X+rect.W > maxX {
			maxX = rect.X + rect.W
		}
		if rect.Y+rect.H > maxY {
			maxY = rect.Y + rect.H
		}
	}
	layout.Width = maxX + padding
	layout.Height = maxY + padding

	return layout
}

// layoutC4Row positions elements in a grid starting at (startX, startY).
// Returns the total width used.
func layoutC4Row(elems []C4Element, startX, startY, elemW, elemH, spacingX float64, perRow int, positioned map[string]c4Positioned) float64 {
	maxW := 0.0
	for i, elem := range elems {
		col := i % perRow
		row := i / perRow
		x := startX + float64(col)*(elemW+spacingX)
		y := startY + float64(row)*(elemH+spacingX)
		positioned[elem.Alias] = c4Positioned{
			elem: elem,
			x:    x, y: y,
			w: elemW, h: elemH,
		}
		if x+elemW > maxW {
			maxW = x + elemW
		}
	}
	return maxW
}

// c4EdgeCenters returns the midpoints of each edge: top, bottom, left, right.
func c4EdgeCenters(pos c4Positioned) [4][2]float64 {
	cx := pos.x + pos.w/2
	cy := pos.y + pos.h/2
	return [4][2]float64{
		{cx, pos.y},          // top
		{cx, pos.y + pos.h},  // bottom
		{pos.x, cy},          // left
		{pos.x + pos.w, cy},  // right
	}
}

// segmentIntersectsRect tests whether line segment (x1,y1)-(x2,y2) intersects
// axis-aligned rectangle (rx,ry,rw,rh) using the Liang-Barsky algorithm.
func segmentIntersectsRect(x1, y1, x2, y2, rx, ry, rw, rh float64) bool {
	dx := x2 - x1
	dy := y2 - y1
	p := [4]float64{-dx, dx, -dy, dy}
	q := [4]float64{x1 - rx, rx + rw - x1, y1 - ry, ry + rh - y1}

	tMin := 0.0
	tMax := 1.0
	for i := range 4 {
		if p[i] == 0 {
			if q[i] < 0 {
				return false // parallel and outside
			}
			continue
		}
		t := q[i] / p[i]
		if p[i] < 0 {
			if t > tMin {
				tMin = t
			}
		} else {
			if t < tMax {
				tMax = t
			}
		}
		if tMin > tMax {
			return false
		}
	}
	return true
}

// c4OptimizeConnections picks the best edge-center pair for each relationship
// by minimizing crossings with other boxes, breaking ties by shortest distance.
func c4OptimizeConnections(rels []C4Rel, positioned map[string]c4Positioned, boundaryRects []LayoutRect) [][4]float64 {
	results := make([][4]float64, len(rels))

	// Build list of obstacle rects (all positioned elements + boundary rects)
	type rect struct{ x, y, w, h float64 }
	obstacles := make([]rect, 0, len(positioned)+len(boundaryRects))
	for _, p := range positioned {
		obstacles = append(obstacles, rect{p.x, p.y, p.w, p.h})
	}
	for _, br := range boundaryRects {
		obstacles = append(obstacles, rect{br.X, br.Y, br.W, br.H})
	}

	for i, rel := range rels {
		from, okFrom := positioned[rel.From]
		to, okTo := positioned[rel.To]
		if !okFrom || !okTo {
			continue
		}

		fromPts := c4EdgeCenters(from)
		toPts := c4EdgeCenters(to)

		bestCrossings := math.MaxInt
		bestDist := math.MaxFloat64
		var best [4]float64

		fromRect := rect{from.x, from.y, from.w, from.h}
		toRect := rect{to.x, to.y, to.w, to.h}

		for _, fp := range fromPts {
			for _, tp := range toPts {
				crossings := 0
				for _, obs := range obstacles {
					// Skip the from and to boxes themselves
					if obs == fromRect || obs == toRect {
						continue
					}
					if segmentIntersectsRect(fp[0], fp[1], tp[0], tp[1], obs.x, obs.y, obs.w, obs.h) {
						crossings++
					}
				}
				dist := math.Hypot(tp[0]-fp[0], tp[1]-fp[1])
				if crossings < bestCrossings || (crossings == bestCrossings && dist < bestDist) {
					bestCrossings = crossings
					bestDist = dist
					best = [4]float64{fp[0], fp[1], tp[0], tp[1]}
				}
			}
		}
		results[i] = best
	}
	return results
}

// addC4Person renders a person shape: circle head on top of a rounded-rect body.
func addC4Person(layout *Layout, x, y, w, h float64, fill string) {
	headR := 20.0
	headCX := x + w/2
	headCY := y + headR

	// Head circle
	layout.Circles = append(layout.Circles, LayoutCircle{
		CX:          headCX,
		CY:          headCY,
		R:           headR,
		Fill:        fill,
		Stroke:      fill,
		StrokeWidth: 1,
	})

	// Body rect below head
	bodyY := y + headR*2 - 4 // slight overlap with head
	bodyH := h - headR*2 + 4
	layout.Rects = append(layout.Rects, LayoutRect{
		X:           x,
		Y:           bodyY,
		W:           w,
		H:           bodyH,
		RX:          8,
		RY:          8,
		Fill:        fill,
		Stroke:      fill,
		StrokeWidth: 1,
	})
}

// addC4Cylinder renders a cylinder (database) shape using paths.
func addC4Cylinder(layout *Layout, x, y, w, h float64, fill string) {
	ellipseH := 12.0
	bodyH := h - ellipseH*2

	// Main body rect
	layout.Rects = append(layout.Rects, LayoutRect{
		X:           x,
		Y:           y + ellipseH,
		W:           w,
		H:           bodyH,
		Fill:        fill,
		Stroke:      fill,
		StrokeWidth: 1,
	})

	// Top ellipse
	cx := x + w/2
	topD := "M " + formatFloat(x) + " " + formatFloat(y+ellipseH) +
		" A " + formatFloat(w/2) + " " + formatFloat(ellipseH) + " 0 1 1 " +
		formatFloat(x+w) + " " + formatFloat(y+ellipseH) +
		" A " + formatFloat(w/2) + " " + formatFloat(ellipseH) + " 0 1 1 " +
		formatFloat(x) + " " + formatFloat(y+ellipseH) + " Z"
	_ = cx
	layout.Paths = append(layout.Paths, LayoutPath{
		D:           topD,
		Fill:        fill,
		Stroke:      fill,
		StrokeWidth: 1,
	})

	// Bottom ellipse
	bottomY := y + h - ellipseH
	bottomD := "M " + formatFloat(x) + " " + formatFloat(bottomY) +
		" A " + formatFloat(w/2) + " " + formatFloat(ellipseH) + " 0 1 0 " +
		formatFloat(x+w) + " " + formatFloat(bottomY) +
		" A " + formatFloat(w/2) + " " + formatFloat(ellipseH) + " 0 1 0 " +
		formatFloat(x) + " " + formatFloat(bottomY) + " Z"
	layout.Paths = append(layout.Paths, LayoutPath{
		D:           bottomD,
		Fill:        fill,
		Stroke:      fill,
		StrokeWidth: 1,
	})
}
