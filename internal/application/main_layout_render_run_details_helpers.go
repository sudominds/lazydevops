package application

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func joinColumnsFixed(left []string, right []string, leftWidth int, rightWidth int, gap int) []string {
	if leftWidth < 1 {
		leftWidth = 1
	}
	if rightWidth < 1 {
		rightWidth = 1
	}
	if gap < 0 {
		gap = 0
	}
	maxRows := len(left)
	if len(right) > maxRows {
		maxRows = len(right)
	}
	rows := make([]string, 0, maxRows)
	for i := 0; i < maxRows; i++ {
		leftLine := ""
		rightLine := ""
		if i < len(left) {
			leftLine = left[i]
		}
		if i < len(right) {
			rightLine = right[i]
		}
		rows = append(rows, padRenderedLineToWidth(leftLine, leftWidth)+strings.Repeat(" ", gap)+padRenderedLineToWidth(rightLine, rightWidth))
	}
	return rows
}

func fitLineToWidth(line string, width int) string {
	if width < 1 {
		return ""
	}
	normalizedLine := strings.ReplaceAll(line, "\r", "")
	normalizedLine = strings.ReplaceAll(normalizedLine, "\t", "    ")
	truncated := truncateWithEllipsis(normalizedLine, width)
	lineWidth := lipgloss.Width(truncated)
	if lineWidth >= width {
		return truncated
	}
	return truncated + strings.Repeat(" ", width-lineWidth)
}

func padRenderedLineToWidth(line string, width int) string {
	if width < 1 {
		return ""
	}
	lineWidth := lipgloss.Width(line)
	if lineWidth >= width {
		return line
	}
	return line + strings.Repeat(" ", width-lineWidth)
}

var ansiEscapeSequencePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func sanitizeDisplayLine(value string) string {
	withoutCarriageReturns := strings.ReplaceAll(value, "\r", "")
	withoutANSI := ansiEscapeSequencePattern.ReplaceAllString(withoutCarriageReturns, "")
	return strings.Map(func(character rune) rune {
		if character < 32 && character != '\n' && character != '\t' {
			return -1
		}
		if character == 127 {
			return -1
		}
		return character
	}, withoutANSI)
}

func (model MainLayoutModel) renderRunDetailsSection(sectionIndex int, width int, height int, lines []string) []string {
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	visibleLines := make([]string, 0, visibleHeight)
	for index := 0; index < visibleHeight && index < len(lines); index++ {
		visibleLines = append(visibleLines, fitLineToWidth(lines[index], contentWidth))
	}
	for len(visibleLines) < visibleHeight {
		visibleLines = append(visibleLines, strings.Repeat(" ", contentWidth))
	}

	panelStyle := model.runDetailsSectionStyle(sectionIndex).
		Padding(0, 1)
	rendered := panelStyle.Render(strings.Join(visibleLines, "\n"))
	renderedLines := strings.Split(rendered, "\n")
	// Keep full rendered panel to avoid clipping border rows.
	for len(renderedLines) < height {
		renderedLines = append(renderedLines, strings.Repeat(" ", width))
	}
	if len(renderedLines) > height {
		renderedLines = renderedLines[:height]
	}
	return renderedLines
}

func (model MainLayoutModel) runDetailsSectionStyle(sectionIndex int) lipgloss.Style {
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(model.palette.Muted))
	if model.runDetailsFocusSection == sectionIndex {
		style = style.BorderForeground(lipgloss.Color(model.palette.Accent))
	} else if sectionIndex == 12 && (model.runDetailsFocusSection == 1 || model.runDetailsFocusSection == 2) {
		style = style.BorderForeground(lipgloss.Color(model.palette.Accent))
	}
	return style
}
