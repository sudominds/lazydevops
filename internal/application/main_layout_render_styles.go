package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"lazydevops/internal/application/components/search"
)

func (model MainLayoutModel) renderSelectableLine(isSelected bool, label string) string {
	if !isSelected {
		return "   " + label
	}

	if model.currentMode == inputModeInsert {
		insertForeground := model.palette.InsertSelectedForeground
		if model.listRainbowSelectionEnabled() {
			insertForeground = model.listRainbowColor(0)
		}
		insertModeSelectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(insertForeground)).
			Background(lipgloss.Color(model.palette.InsertSelectedBackground))
		return insertModeSelectionStyle.Render(" > " + label)
	}

	selectedForeground := model.palette.SelectedForeground
	if model.listRainbowSelectionEnabled() {
		selectedForeground = model.listRainbowColor(0)
	}
	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(selectedForeground)).
		Background(lipgloss.Color(model.palette.SelectedBackground))
	return selectedStyle.Render(" > " + label)
}

func (model MainLayoutModel) renderSetSelectableLine(isSelected bool, label string, query string, colorIndex int) string {
	prefix := "   "
	highlightedLabel := model.highlightSearchMatches(label, query)
	if isSelected {
		selectedForeground := model.palette.SelectedForeground
		if model.listRainbowSelectionEnabled() {
			selectedForeground = model.listRainbowColor(colorIndex)
		}
		selectedStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(selectedForeground)).
			Background(lipgloss.Color(model.palette.SelectedBackground))
		// Render full-row selection without nested ANSI highlight resets.
		return selectedStyle.Render(" > " + sanitizeDisplayLine(highlightedLabel))
	}
	return prefix + highlightedLabel
}

func (model MainLayoutModel) highlightSearchMatches(value string, query string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	matchIndexes := search.MatchRuneIndexes(value, query)
	if len(matchIndexes) == 0 {
		return value
	}

	var builder strings.Builder
	matchCounter := 0
	for runeIndex, character := range runes {
		if matchIndexes[runeIndex] {
			builder.WriteString(model.renderSearchMatchRune(character, matchCounter))
			matchCounter++
			continue
		}
		builder.WriteRune(character)
	}
	return builder.String()
}

func (model MainLayoutModel) renderSearchMatchRune(character rune, matchIndex int) string {
	matchStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Accent))
	if strings.EqualFold(strings.TrimSpace(model.searchSettings.MatchHighlightMode), "rainbow") && len(model.searchSettings.RainbowColors) > 0 {
		colorIndex := matchIndex % len(model.searchSettings.RainbowColors)
		matchStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.searchSettings.RainbowColors[colorIndex]))
	}
	return matchStyle.Render(string(character))
}

func (model MainLayoutModel) listHighlightMode() string {
	mode := strings.ToLower(strings.TrimSpace(model.mainLayoutSettings.ListHighlightMode))
	switch mode {
	case "accent", "selection":
		return mode
	default:
		return "off"
	}
}

func (model MainLayoutModel) listRainbowAccentEnabled() bool {
	return model.listHighlightMode() == "accent" && len(model.searchSettings.RainbowColors) > 0
}

func (model MainLayoutModel) listRainbowSelectionEnabled() bool {
	return model.listHighlightMode() == "selection" && len(model.searchSettings.RainbowColors) > 0
}

func (model MainLayoutModel) listRainbowColor(index int) string {
	if len(model.searchSettings.RainbowColors) == 0 {
		return model.palette.Accent
	}
	if index < 0 {
		index = 0
	}
	return model.searchSettings.RainbowColors[index%len(model.searchSettings.RainbowColors)]
}

func (model MainLayoutModel) renderSelectableCardLines(contentWidth int, isSelected bool, lines []listCardLine, searchQuery string, colorIndex int) []string {
	if contentWidth < 18 {
		contentWidth = 18
	}

	marker := "  "
	if isSelected {
		markerColor := model.palette.Accent
		if model.listRainbowAccentEnabled() {
			markerColor = model.listRainbowColor(colorIndex)
		}
		marker = lipgloss.NewStyle().Foreground(lipgloss.Color(markerColor)).Render("▌ ")
	}

	innerWidth := contentWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	renderedLines := make([]string, 0, len(lines))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))
	for _, cardLine := range lines {
		lineText := strings.TrimSpace(cardLine.Text)
		if lineText == "" {
			continue
		}
		highlightedLine := model.highlightSearchMatches(lineText, searchQuery)
		fittedLine := fitLineToWidth(highlightedLine, innerWidth)
		if cardLine.Muted {
			fittedLine = metaStyle.Render(fittedLine)
		}
		renderedLines = append(renderedLines, marker+fittedLine)
	}
	if len(renderedLines) == 0 {
		renderedLines = append(renderedLines, marker)
	}

	rowStyle := lipgloss.NewStyle().Width(contentWidth)
	return strings.Split(rowStyle.Render(strings.Join(renderedLines, "\n")), "\n")
}

func (model MainLayoutModel) renderStatusBar(width int) string {
	modeLabel := " NORMAL "
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(model.palette.ModeNormalForeground)).
		Background(lipgloss.Color(model.palette.ModeNormalBackground))
	if model.currentMode == inputModeInsert {
		modeLabel = " INSERT "
		modeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(model.palette.ModeInsertForeground)).
			Background(lipgloss.Color(model.palette.ModeInsertBackground))
	}

	leftSegment := modeStyle.Render(modeLabel)
	organizationText := fmt.Sprintf("Org: %s", model.organizationURL)
	leftText := organizationText
	rightText := ""
	if model.mainLayoutSettings.ShowPathInStatusBar {
		pathText := model.breadcrumb()
		if model.mainLayoutSettings.StatusBarPathSide == "left" {
			leftText = pathText
			rightText = organizationText
		} else {
			leftText = organizationText
			rightText = pathText
		}
	}

	leftInfoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.StatusBarSecondaryForeground))
	modeAndGapWidth := lipgloss.Width(leftSegment) + 1
	minimumRightWidth := 1
	if strings.TrimSpace(rightText) != "" {
		minimumRightWidth = 18
	}
	maxLeftInfoWidth := width - modeAndGapWidth - minimumRightWidth
	if maxLeftInfoWidth < 1 {
		maxLeftInfoWidth = 1
	}
	leftInfoSegment := leftInfoStyle.Render(truncateWithEllipsis(leftText, maxLeftInfoWidth))
	rightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.StatusBarSecondaryForeground))

	leftContentSegment := leftSegment + " " + leftInfoSegment

	maxRightWidth := width - lipgloss.Width(leftContentSegment) - 1
	if maxRightWidth < 1 {
		maxRightWidth = 1
	}
	rightText = truncateWithEllipsis(rightText, maxRightWidth)

	rightSegment := rightStyle.Render(rightText)
	spacerWidth := width - lipgloss.Width(leftContentSegment) - lipgloss.Width(rightSegment)
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	spacerSegment := strings.Repeat(" ", spacerWidth)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(model.palette.StatusBarForeground)).
		Background(lipgloss.Color(model.palette.StatusBarBackground))

	return statusBarStyle.Width(width).Render(leftContentSegment + spacerSegment + rightSegment)
}

func truncateWithEllipsis(value string, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	if maxWidth <= 3 {
		clipped := ""
		for _, character := range value {
			next := clipped + string(character)
			if lipgloss.Width(next) > maxWidth {
				break
			}
			clipped = next
		}
		return clipped
	}

	allowedWidth := maxWidth - 3
	clipped := ""
	for _, character := range value {
		next := clipped + string(character)
		if lipgloss.Width(next) > allowedWidth {
			break
		}
		clipped = next
	}
	return clipped + "..."
}
