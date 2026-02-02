package app

type Viewport struct {
	Layout       *Layout
	Config       LayoutConfig
	ScrollOffset int
	ScreenHeight int

	VisibleStart int
	VisibleEnd   int
	HasMoreAbove bool
	HasMoreBelow bool

	itemRowStart []int // Starting row for each visible layout item
}

func NewViewport(layout *Layout, screenHeight int, config LayoutConfig) *Viewport {
	return &Viewport{
		Layout:       layout,
		Config:       config,
		ScreenHeight: screenHeight,
	}
}

func (v *Viewport) availableHeight() int {
	if v.ScreenHeight <= v.Config.FooterHeight {
		return 1
	}
	return v.ScreenHeight - v.Config.FooterHeight
}

func (v *Viewport) contentHeight(reserveMoreBelow bool) int {
	availHeight := v.availableHeight()
	if v.ScrollOffset > 0 {
		availHeight-- // "more above" indicator
	}
	if reserveMoreBelow {
		availHeight-- // "more below" indicator
	}
	if availHeight < 1 {
		availHeight = 1
	}
	return availHeight
}

func (v *Viewport) ComputeVisibility(scrollOffset int) {
	v.ScrollOffset = scrollOffset
	v.HasMoreAbove = scrollOffset > 0
	v.HasMoreBelow = false
	v.VisibleStart = -1
	v.VisibleEnd = -1
	v.itemRowStart = nil

	if v.Layout == nil || len(v.Layout.Items) == 0 {
		return
	}

	// When scrollOffset is 0, always start from the beginning
	// This ensures the project title is always visible at startup
	if scrollOffset == 0 {
		v.VisibleStart = 0
	} else {
		// Find the starting layout item index from scrollOffset
		// ScrollOffset refers to position indices (selectable items), not layout items
		positionCursor := 0
		for i, item := range v.Layout.Items {
			if item.PositionIndex >= 0 && positionCursor >= scrollOffset {
				v.VisibleStart = i
				break
			}
			if item.PositionIndex >= 0 {
				positionCursor++
			}
		}
		if v.VisibleStart < 0 {
			v.VisibleStart = 0
		}
	}

	// Two-pass approach: first try without reserving "more below" space
	// If content doesn't fit, recalculate with the reserved space
	v.computeVisibleRange(false)

	if v.HasMoreBelow {
		// Content didn't fit, recalculate with reserved indicator space
		v.computeVisibleRange(true)
	}
}

func (v *Viewport) computeVisibleRange(reserveMoreBelow bool) {
	availHeight := v.contentHeight(reserveMoreBelow)
	usedHeight := 0

	v.HasMoreBelow = false
	v.VisibleEnd = -1
	v.itemRowStart = nil

	startRow := 0
	if v.HasMoreAbove {
		startRow = 1
	}

	for i := v.VisibleStart; i < len(v.Layout.Items); i++ {
		item := v.Layout.Items[i]

		if usedHeight+item.Height > availHeight {
			v.HasMoreBelow = true
			v.VisibleEnd = i
			break
		}

		v.itemRowStart = append(v.itemRowStart, startRow+usedHeight)
		usedHeight += item.Height
		v.VisibleEnd = i + 1
	}

	// Check if there are more items beyond what we rendered
	if v.VisibleEnd < len(v.Layout.Items) {
		v.HasMoreBelow = true
	}
}

func (v *Viewport) EnsureVisible(posIndex int) int {
	if v.Layout == nil || len(v.Layout.Items) == 0 || posIndex < 0 {
		return 0
	}

	// Find the layout item for this position index
	targetItemIdx := -1
	for i, item := range v.Layout.Items {
		if item.PositionIndex == posIndex {
			targetItemIdx = i
			break
		}
	}

	if targetItemIdx < 0 {
		return v.ScrollOffset
	}

	scrollOffset := v.ScrollOffset

	// If selected position is before scroll offset, scroll up
	if posIndex < scrollOffset {
		return posIndex
	}

	// Calculate if selected is visible
	v.ComputeVisibility(scrollOffset)

	// Check if target item is fully visible
	for i := v.VisibleStart; i < v.VisibleEnd; i++ {
		if i == targetItemIdx {
			// Target is visible
			return scrollOffset
		}
	}

	// Target is not visible, need to scroll down
	// Increment scroll offset until target becomes visible
	for scrollOffset < posIndex {
		scrollOffset++
		v.ComputeVisibility(scrollOffset)

		for i := v.VisibleStart; i < v.VisibleEnd; i++ {
			if i == targetItemIdx {
				return scrollOffset
			}
		}
	}

	return scrollOffset
}

func (v *Viewport) RowToPosition(row int) int {
	if v.Layout == nil || v.VisibleStart < 0 || v.VisibleEnd <= v.VisibleStart {
		return -1
	}

	for i := 0; i < len(v.itemRowStart); i++ {
		itemIdx := v.VisibleStart + i
		if itemIdx >= len(v.Layout.Items) {
			break
		}

		item := v.Layout.Items[itemIdx]
		itemStart := v.itemRowStart[i]
		itemEnd := itemStart + item.Height

		if row >= itemStart && row < itemEnd {
			return item.PositionIndex
		}
	}

	return -1
}

func (v *Viewport) CenterOnPosition(posIndex int) int {
	if v.Layout == nil || len(v.Layout.Items) == 0 || posIndex < 0 {
		return 0
	}

	// Find the layout item for this position
	targetItemIdx := -1
	var targetItem LayoutItem
	for i, item := range v.Layout.Items {
		if item.PositionIndex == posIndex {
			targetItemIdx = i
			targetItem = item
			break
		}
	}

	if targetItemIdx < 0 {
		return 0
	}

	availHeight := v.availableHeight()
	availHeight -= 2 // Reserve for scroll indicators
	if availHeight < 1 {
		availHeight = 1
	}

	targetSpaceAbove := (availHeight - targetItem.Height) / 2
	if targetSpaceAbove < 0 {
		targetSpaceAbove = 0
	}

	// Calculate cumulative heights to find ideal scroll offset
	usedHeight := 0
	scrollOffset := 0

	for i := 0; i < targetItemIdx; i++ {
		item := v.Layout.Items[i]
		if item.PositionIndex < 0 {
			usedHeight += item.Height
			continue
		}

		newHeight := usedHeight + item.Height
		if newHeight > targetSpaceAbove && i > 0 {
			scrollOffset = item.PositionIndex
			usedHeight = item.Height
		} else {
			usedHeight = newHeight
		}
	}

	if scrollOffset < 0 {
		scrollOffset = 0
	}
	if scrollOffset > posIndex {
		scrollOffset = posIndex
	}

	return scrollOffset
}
