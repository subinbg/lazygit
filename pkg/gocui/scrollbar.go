package gocui

import "math"

// returns start and height of scrollbar
// `max` is the maximum possible value of `position`
func calcScrollbar(listSize int, pageSize int, position int, scrollAreaSize int) (int, int) {
	height := calcScrollbarHeight(listSize, pageSize, scrollAreaSize)
	// assume we can't scroll past the last item
	maxPosition := listSize - pageSize
	if maxPosition <= 0 {
		return 0, height
	}
	if position == maxPosition {
		return scrollAreaSize - height, height
	}
	// we only want to show the scrollbar at the top or bottom positions if we're at the end. Hence the .Ceil (for moving the scrollbar once we scroll down) and the -1 (for pretending there's a smaller range than we actually have, with the above condition ensuring we snap to the bottom once we're at the end of the list)
	start := int(math.Ceil(((float64(position) / float64(maxPosition)) * float64(scrollAreaSize-height-1))))
	return start, height
}

func calcScrollbarHeight(listSize int, pageSize int, scrollAreaSize int) int {
	if pageSize >= listSize {
		return scrollAreaSize
	}

	return int((float64(pageSize) / float64(listSize)) * float64(scrollAreaSize))
}

// scrollbarDrag is a drag of a view's scrollbar thumb with the mouse.
type scrollbarDrag struct {
	view *View
	// rows from the top of the thumb to the row it was grabbed at
	grab int
}

// isOnScrollbar reports whether the given screen position is
// on the view's scrollbar,
// which takes the right edge of its frame
// while its content is taller than the view.
func (v *View) isOnScrollbar(x, y int) bool {
	if !v.Frame || x != v.x1 || y <= v.y0 || y >= v.y1 {
		return false
	}
	showScrollbar, _, _ := calcRealScrollbarStartEnd(v)
	return showScrollbar
}

// startScrollbarDrag starts dragging the view's scrollbar from a press
// at the given screen row:
// on the thumb it takes hold of the thumb where it was pressed;
// elsewhere it moves the thumb's middle to the row first.
func (g *Gui) startScrollbarDrag(v *View, y int) {
	_, thumbStart, thumbEnd := calcRealScrollbarStartEnd(v)
	grab := (thumbEnd - thumbStart) / 2
	if y >= thumbStart && y <= thumbEnd {
		grab = y - thumbStart
	}
	g.scrollbarDrag = &scrollbarDrag{view: v, grab: grab}
	g.captureMouse(v)
	g.dragScrollbarTo(y)
}

// dragScrollbarTo scrolls the dragged scrollbar's view
// so that the thumb's grabbed row is at the given screen row,
// as far as the thumb can go.
func (g *Gui) dragScrollbarTo(y int) {
	v := g.scrollbarDrag.view
	height := v.InnerHeight()
	fullHeight := v.scrollbarContentHeight() - v.scrollMargin()
	if v.CanScrollPastBottom {
		fullHeight += height
	}
	maxPosition := fullHeight - height
	if height < 2 || maxPosition <= 0 {
		return
	}

	scrollAreaSize := height - 1
	thumbHeight := calcScrollbarHeight(fullHeight, height, scrollAreaSize)
	// the rows the top of the thumb can be at, from the top of the view on
	track := scrollAreaSize - thumbHeight
	if track <= 0 {
		return
	}
	thumbStart := max(0, min(track, y-g.scrollbarDrag.grab-(v.y0+1)))
	position := int(math.Round(float64(thumbStart) * float64(maxPosition) / float64(track)))

	if delta := position - v.OriginY(); delta < 0 {
		v.ScrollUp(-delta)
	} else if delta > 0 {
		v.ScrollDown(delta)
	}
	if g.OnScrollbarScroll != nil {
		g.OnScrollbarScroll(v)
	}
}
