package app

import (
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/app/components"
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type LayoutItemKind int

const (
	LayoutProject LayoutItemKind = iota
	LayoutCategory
	LayoutTask
	LayoutSeparator     // In-category divider row (domain.KindSeparator)
	LayoutDescription   // Inline description block for the task above it (own focusable row)
	LayoutEmptyCategory // "(no tasks)" placeholder
	LayoutFolded        // "(folded)" placeholder
	LayoutSpacing       // Blank lines between elements
)

type LayoutItem struct {
	Kind          LayoutItemKind
	Height        int // Screen rows this item occupies
	PositionIndex int // Index into model.positions (-1 for non-selectable)
	CategoryIndex int
	TaskIndex     int
}

type Layout struct {
	Items       []LayoutItem
	TotalHeight int
}

type LayoutConfig struct {
	FooterHeight      int
	BlankAfterProject int
	BlankBetweenCats  int
	BlankAfterCatHead int
}

func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		FooterHeight:      footerHeight,
		BlankAfterProject: blankAfterProj,
		BlankBetweenCats:  blankBetweenCat,
		BlankAfterCatHead: blankAfterCat,
	}
}

type LayoutBuilder struct {
	config             LayoutConfig
	width              int
	statusDisplay      string
	filter             *FilterState
	fold               *FoldState
	expandDescriptions bool
	// cursorCat/cursorTask address the description row the cursor sits on
	// (-1/-1 when the cursor is elsewhere). That row renders in full; every
	// other description row is capped to its preview height.
	cursorCat  int
	cursorTask int
}

func NewLayoutBuilder(config LayoutConfig, width int, statusDisplay string, filter *FilterState, fold *FoldState) *LayoutBuilder {
	return &LayoutBuilder{
		config:        config,
		width:         width,
		statusDisplay: statusDisplay,
		filter:        filter,
		fold:          fold,
		cursorCat:     -1,
		cursorTask:    -1,
	}
}

func (b *LayoutBuilder) WithExpandedDescriptions(v bool) *LayoutBuilder {
	b.expandDescriptions = v
	return b
}

func (b *LayoutBuilder) WithCursorDescription(catIdx, taskIdx int) *LayoutBuilder {
	b.cursorCat = catIdx
	b.cursorTask = taskIdx
	return b
}

func (b *LayoutBuilder) Build(project domain.Project, positions []selection.Position) Layout {
	var items []LayoutItem
	totalHeight := 0
	posIndex := 0

	// Project line (first focusable item)
	projectHeight := 1 // Project line doesn't wrap
	items = append(items, LayoutItem{
		Kind:          LayoutProject,
		Height:        projectHeight,
		PositionIndex: posIndex,
		CategoryIndex: -1,
		TaskIndex:     -1,
	})
	totalHeight += projectHeight
	posIndex++

	// Spacing after project
	if b.config.BlankAfterProject > 0 {
		items = append(items, LayoutItem{
			Kind:          LayoutSpacing,
			Height:        b.config.BlankAfterProject,
			PositionIndex: -1,
			CategoryIndex: -1,
			TaskIndex:     -1,
		})
		totalHeight += b.config.BlankAfterProject
	}

	for catIdx, category := range project.Categories {
		// Spacing between categories (not before first)
		if catIdx > 0 && b.config.BlankBetweenCats > 0 {
			items = append(items, LayoutItem{
				Kind:          LayoutSpacing,
				Height:        b.config.BlankBetweenCats,
				PositionIndex: -1,
				CategoryIndex: -1,
				TaskIndex:     -1,
			})
			totalHeight += b.config.BlankBetweenCats
		}

		// Category header (add extra width for fold indicator + status badge)
		catSuffixWidth := 0
		if category.AggregateStatus() != "" {
			catSuffixWidth += 4 // " [x]"
		}
		catHeight := countWrappedLines(category.Name, b.width, prefixWidth+2+catSuffixWidth)
		items = append(items, LayoutItem{
			Kind:          LayoutCategory,
			Height:        catHeight,
			PositionIndex: posIndex,
			CategoryIndex: catIdx,
			TaskIndex:     -1,
		})
		totalHeight += catHeight
		posIndex++

		isFolded := b.fold != nil && b.fold.IsFolded(category.ID)
		if isFolded {
			items = append(items, LayoutItem{
				Kind:          LayoutFolded,
				Height:        1,
				PositionIndex: -1,
				CategoryIndex: catIdx,
				TaskIndex:     -1,
			})
			totalHeight++
			continue
		}

		visibleTaskCount := 0
		for _, task := range category.Tasks {
			if taskRowVisible(b.filter, task, category.ID) {
				visibleTaskCount++
			}
		}

		if visibleTaskCount == 0 {
			// "(no tasks)" placeholder - not selectable
			items = append(items, LayoutItem{
				Kind:          LayoutEmptyCategory,
				Height:        1,
				PositionIndex: -1,
				CategoryIndex: catIdx,
				TaskIndex:     -1,
			})
			totalHeight++
			continue
		}

		// Spacing after category header (before tasks)
		if b.config.BlankAfterCatHead > 0 {
			items = append(items, LayoutItem{
				Kind:          LayoutSpacing,
				Height:        b.config.BlankAfterCatHead,
				PositionIndex: -1,
				CategoryIndex: -1,
				TaskIndex:     -1,
			})
			totalHeight += b.config.BlankAfterCatHead
		}

		// Tasks (consecutive tasks have no blank lines between them)
		for taskIdx, task := range category.Tasks {
			if !taskRowVisible(b.filter, task, category.ID) {
				continue
			}
			if task.IsSeparator() {
				items = append(items, LayoutItem{
					Kind:          LayoutSeparator,
					Height:        1,
					PositionIndex: posIndex,
					CategoryIndex: catIdx,
					TaskIndex:     taskIdx,
				})
				totalHeight++
				posIndex++
				continue
			}
			taskHeight := b.countTaskLines(task)
			items = append(items, LayoutItem{
				Kind:          LayoutTask,
				Height:        taskHeight,
				PositionIndex: posIndex,
				CategoryIndex: catIdx,
				TaskIndex:     taskIdx,
			})
			totalHeight += taskHeight
			posIndex++

			if b.expandDescriptions && task.Description != "" {
				full := catIdx == b.cursorCat && taskIdx == b.cursorTask
				descHeight := components.DescriptionHeight(task.Description, b.width, taskTitleColumn(task, b.statusDisplay), full)
				items = append(items, LayoutItem{
					Kind:          LayoutDescription,
					Height:        descHeight,
					PositionIndex: posIndex,
					CategoryIndex: catIdx,
					TaskIndex:     taskIdx,
				})
				totalHeight += descHeight
				posIndex++
			}
		}
	}

	return Layout{
		Items:       items,
		TotalHeight: totalHeight,
	}
}

func (b *LayoutBuilder) countTaskLines(task domain.Task) int {
	if b.width <= 0 {
		return 1
	}
	prefix := "  "
	priorityIcon := ui.PriorityIcon(task.Priority)
	statusText := statusLabel(task.Status, b.statusDisplay)
	iconText := ""
	if priorityIcon != "" {
		iconText = priorityIcon + " "
	}
	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	return countWrappedLines(task.Title, b.width, overhead)
}

func (m *model) buildLayout() *Layout {
	w := m.ui.Screen.Width
	cursorCat, cursorTask := -1, -1
	if pos, ok := m.selectedPosition(); ok && pos.Kind == selection.FocusDescription {
		cursorCat, cursorTask = pos.CategoryIndex, pos.TaskIndex
	}
	c := m.ui.layout
	if c.layout != nil && c.width == w && c.cursorCat == cursorCat && c.cursorTask == cursorTask {
		return c.layout
	}
	builder := NewLayoutBuilder(m.layoutConfig(), w, m.deps.CfgManager.Get().StatusDisplay, &m.ui.Filter, &m.ui.Fold).
		WithExpandedDescriptions(m.ui.Screen.ExpandDescriptions).
		WithCursorDescription(cursorCat, cursorTask)
	layout := builder.Build(m.project, m.positions())
	m.ui.layout = layoutCache{width: w, cursorCat: cursorCat, cursorTask: cursorTask, layout: &layout}
	return &layout
}
