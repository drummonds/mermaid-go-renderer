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

		if isDbKind(elem.Kind) {
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
		lineY := pos.y + 30
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

	// Render relationships
	for _, rel := range graph.C4Rels {
		from, okFrom := positioned[rel.From]
		to, okTo := positioned[rel.To]
		if !okFrom || !okTo {
			continue
		}

		x1, y1 := c4ConnPoint(from, to)
		x2, y2 := c4ConnPoint(to, from)

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

// c4ConnPoint computes the connection point on the border of 'from' facing 'to'.
func c4ConnPoint(from, to c4Positioned) (float64, float64) {
	fcx := from.x + from.w/2
	fcy := from.y + from.h/2
	tcx := to.x + to.w/2
	tcy := to.y + to.h/2

	dx := tcx - fcx
	dy := tcy - fcy
	if dx == 0 && dy == 0 {
		return fcx, fcy
	}

	// Determine which edge of the rect to use
	absDx := math.Abs(dx)
	absDy := math.Abs(dy)

	if absDx/from.w > absDy/from.h {
		// Horizontal edge
		if dx > 0 {
			return from.x + from.w, fcy + dy*(from.w/2)/absDx
		}
		return from.x, fcy - dy*(from.w/2)/absDx
	}
	// Vertical edge
	if dy > 0 {
		return fcx + dx*(from.h/2)/absDy, from.y + from.h
	}
	return fcx - dx*(from.h/2)/absDy, from.y
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
