package app

import (
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type LayoutItemKind int

const (
	LayoutProject LayoutItemKind = iota
	LayoutCategory
	LayoutTask
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
	config        LayoutConfig
	width         int
	statusDisplay string
	filter        *FilterState
	fold          *FoldState
}

func NewLayoutBuilder(config LayoutConfig, width int, statusDisplay string, filter *FilterState, fold *FoldState) *LayoutBuilder {
	return &LayoutBuilder{
		config:        config,
		width:         width,
		statusDisplay: statusDisplay,
		filter:        filter,
		fold:          fold,
	}
}

func (b *LayoutBuilder) Build(project domain.Project, positions []focusPosition) Layout {
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

		// Category header (add extra width for fold indicator)
		catHeight := countWrappedLines(category.Name, b.width, prefixWidth+2)
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
			if b.filter == nil || b.filter.IsStatusVisible(task.Status) {
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
			if b.filter != nil && !b.filter.IsStatusVisible(task.Status) {
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
	builder := NewLayoutBuilder(DefaultLayoutConfig(), m.ui.Width, m.deps.CfgManager.Get().StatusDisplay, &m.ui.Filter, &m.ui.Fold)
	layout := builder.Build(m.project, m.positions())
	return &layout
}
