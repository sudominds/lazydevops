package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (model MainLayoutModel) renderRunDetailsViewportLines(contentLineBudget int) []string {
	allLines := model.renderRunDetailsLines()
	if contentLineBudget < 1 {
		return []string{}
	}
	if len(allLines) <= contentLineBudget {
		return allLines
	}

	maxOffset := len(allLines) - contentLineBudget
	offset := model.runDetailsScrollOffset
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	visibleLines := allLines[offset : offset+contentLineBudget]
	return visibleLines
}

func (model MainLayoutModel) renderRunDetailsSplitLines(contentWidth int, contentLineBudget int) []string {
	if !model.hasRunDetails {
		return []string{"No run details loaded.", "Press esc to go back to runs."}
	}
	if contentLineBudget < 1 {
		return []string{}
	}

	summaryLines := model.renderRunSummaryLinesCompact()
	summaryPanelLines := make([]string, 0, len(summaryLines)+1)
	summaryPanelLines = append(summaryPanelLines, "Run Details")
	summaryPanelLines = append(summaryPanelLines, "")
	for _, line := range summaryLines {
		summaryPanelLines = append(summaryPanelLines, truncateWithEllipsis(line, contentWidth-4))
	}

	topPanelHeight := len(summaryPanelLines) + 2
	if topPanelHeight < 8 {
		topPanelHeight = 8
	}
	if contentLineBudget < 18 && topPanelHeight > 8 {
		topPanelHeight = 8
	}
	maxTopPanelHeight := contentLineBudget - 8
	if maxTopPanelHeight < 6 {
		maxTopPanelHeight = 6
	}
	if topPanelHeight > maxTopPanelHeight {
		topPanelHeight = maxTopPanelHeight
	}
	if topPanelHeight < 6 {
		topPanelHeight = 6
	}

	gapRows := 1
	availableSplitLines := contentLineBudget - topPanelHeight - gapRows
	if availableSplitLines < 6 {
		availableSplitLines = 6
	}

	leftWidth := int(float64(contentWidth) * 0.27)
	if leftWidth < 28 {
		leftWidth = 28
	}
	if leftWidth > contentWidth-32 {
		leftWidth = contentWidth - 32
	}
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := contentWidth - leftWidth - 1
	if rightWidth < 20 {
		rightWidth = 20
	}

	topPanelLines := model.renderRunDetailsSection(-1, contentWidth, topPanelHeight, summaryPanelLines)
	leftPanelLines := model.renderRunDetailsTreePanel(leftWidth, availableSplitLines)
	rightPanelLines := model.renderRunDetailsLogsAndContentPanel(rightWidth, availableSplitLines)
	bottomSplitLines := joinColumnsFixed(leftPanelLines, rightPanelLines, leftWidth, rightWidth, 1)

	lines := make([]string, 0, contentLineBudget)
	lines = append(lines, topPanelLines...)
	lines = append(lines, "")
	lines = append(lines, bottomSplitLines...)
	if len(lines) > contentLineBudget {
		lines = lines[:contentLineBudget]
	}
	for len(lines) < contentLineBudget {
		lines = append(lines, "")
	}
	return lines
}

func (model MainLayoutModel) renderRunSummaryLinesCompact() []string {
	runDetails := model.selectedRunDetails
	pipelineName := strings.TrimSpace(runDetails.Pipeline.Name)
	if pipelineName == "" {
		pipelineName = strings.TrimSpace(runDetails.Definition.Name)
	}
	if pipelineName == "" {
		pipelineName = "N/A"
	}
	status := strings.TrimSpace(runDetails.Result)
	if status == "" {
		status = strings.TrimSpace(runDetails.State)
	}
	status = model.statusDisplay(status)
	branch := formatRunBranch(runDetails.SourceBranch)
	if branch == "" {
		branch = "N/A"
	}
	reason := strings.TrimSpace(runDetails.Reason)
	if reason == "" {
		reason = "N/A"
	}
	startedAtRaw := firstNonEmpty(runDetails.StartTime, runDetails.QueueTime, runDetails.CreatedDate)
	finishedAtRaw := strings.TrimSpace(runDetails.FinishedAt)
	duration := formatRunDurationFromBounds(startedAtRaw, finishedAtRaw)
	if duration == "" {
		duration = "N/A"
	}
	startedAt := formatRunDateTime(startedAtRaw)
	if startedAt == "" {
		startedAt = "N/A"
	}
	commitValue := shortCommitHash(runDetails.SourceVersion)
	if commitValue == "" {
		commitValue = "N/A"
	}
	return []string{
		fmt.Sprintf("Run %d | %s", runDetails.ID, pipelineName),
		fmt.Sprintf("Name: %s", configuredValue(runDetails.Name)),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("Trigger Reason: %s", reason),
		fmt.Sprintf("Branch: %s | Commit: %s", branch, commitValue),
		fmt.Sprintf("Started: %s | Duration: %s", startedAt, duration),
	}
}

func (model MainLayoutModel) renderRunDetailsTreePanel(width int, height int) []string {
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	groups := model.buildRunDetailsStageGroups()
	treeItems := model.flattenRunDetailsTree(groups)
	lines := []string{
		fitLineToWidth("Stages/Jobs/Tasks [0]", contentWidth),
		fitLineToWidth("", contentWidth),
	}
	if len(treeItems) == 0 {
		lines = append(lines, fitLineToWidth("No execution stages found.", contentWidth))
		return model.renderRunDetailsSection(0, width, height, lines)
	}

	contentRows := visibleHeight - len(lines)
	if contentRows < 1 {
		return model.renderRunDetailsSection(0, width, height, lines)
	}
	selectedIndex := clampSelection(model.runDetailsTreeCursor, len(treeItems))
	start := selectedIndex - (visibleHeight / 2)
	if start < 0 {
		start = 0
	}
	if start > len(treeItems)-contentRows {
		start = len(treeItems) - contentRows
	}
	if start < 0 {
		start = 0
	}
	end := start + contentRows
	if end > len(treeItems) {
		end = len(treeItems)
	}

	for idx := start; idx < end; idx++ {
		item := treeItems[idx]
		isSelected := idx == selectedIndex
		isFocused := model.runDetailsFocusSection == 0
		line := ""
		if item.Kind == runDetailsTreeItemStage {
			expanded := model.runDetailsExpandedStageIDs[item.StageKey]
			marker := "▸"
			if expanded {
				marker = "▾"
			}
			stageName := model.statusPrefix(item.Stage.Status, item.Stage.Name)
			line = fmt.Sprintf("%s %s (%s)", marker, stageName, item.Stage.Duration)
		} else if item.Kind == runDetailsTreeItemJob {
			expanded := model.isRunDetailsJobExpanded(item.StageKey, item.Job.ID)
			marker := "▸"
			if expanded {
				marker = "▾"
			}
			attemptSuffix := ""
			if strings.TrimSpace(item.Job.Attempt) != "" {
				attemptSuffix = " | attempt:" + item.Job.Attempt
			}
			jobName := model.statusPrefix(item.Job.Status, item.Job.Name)
			line = fmt.Sprintf("  %s %s (%s)%s", marker, jobName, item.Job.Duration, attemptSuffix)
		} else {
			attemptSuffix := ""
			if strings.TrimSpace(item.Task.Attempt) != "" {
				attemptSuffix = " | attempt:" + item.Task.Attempt
			}
			taskName := model.statusPrefix(item.Task.Status, item.Task.Name)
			line = fmt.Sprintf("    - %s (%s)%s", taskName, item.Task.Duration, attemptSuffix)
		}
		lines = append(lines, model.renderRunDetailsTreeRow(line, isSelected, isFocused, contentWidth))
	}

	return model.renderRunDetailsSection(0, width, height, lines)
}

func (model MainLayoutModel) renderRunDetailsTreeRow(value string, isSelected bool, isFocused bool, contentWidth int) string {
	prefix := "  "
	if isSelected {
		prefix = "› "
	}
	if isSelected && isFocused {
		prefix = "> "
	}

	rowValue := value
	if isSelected {
		// ANSI-styled status icons include reset codes that can break full-row background highlights.
		rowValue = sanitizeDisplayLine(value)
	}
	line := fitLineToWidth(prefix+rowValue, contentWidth)
	if !isSelected {
		return line
	}

	if isFocused {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(model.palette.SelectedForeground)).
			Background(lipgloss.Color(model.palette.SelectedBackground)).
			Render(line)
	}

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(model.palette.Accent)).
		Render(line)
}

func (model MainLayoutModel) renderRunDetailsLogsAndContentPanel(width int, height int) []string {
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	lines := []string{
		fitLineToWidth("Log Content [2]", contentWidth),
		fitLineToWidth("", contentWidth),
	}
	if strings.TrimSpace(model.selectedLogID) != "" {
		lines = append(lines, fitLineToWidth("log id: "+strings.TrimSpace(model.selectedLogID), contentWidth))
		lines = append(lines, fitLineToWidth(strings.Repeat("─", contentWidth), contentWidth))
	}

	contentRows := visibleHeight - len(lines)
	if contentRows < 1 {
		contentRows = 1
	}
	if model.isLogContentLoading {
		lines = append(lines, fitLineToWidth("Loading log content...", contentWidth))
		return model.renderRunDetailsSection(12, width, height, lines)
	}
	if model.selectedLogError != "" {
		lines = append(lines, fitLineToWidth("Error: "+model.selectedLogError, contentWidth))
		return model.renderRunDetailsSection(12, width, height, lines)
	}
	if strings.TrimSpace(model.selectedLogContent) == "" {
		lines = append(lines, fitLineToWidth("Select a task/job and press l (or enter) to load logs.", contentWidth))
		return model.renderRunDetailsSection(12, width, height, lines)
	}

	contentLines := strings.Split(model.selectedLogContent, "\n")
	maxOffset := len(contentLines) - contentRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := model.runDetailsScrollOffset
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	endOffset := offset + contentRows
	if endOffset > len(contentLines) {
		endOffset = len(contentLines)
	}
	lines = append(lines, model.renderRunLogContentWindowLines(contentLines, offset, endOffset, contentWidth)...)

	return model.renderRunDetailsSection(12, width, height, lines)
}

func (model MainLayoutModel) renderRunDetailsLogsPanel(width int, height int) []string {
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	logs := model.currentRunDetailsLogs()
	lines := []string{
		fitLineToWidth("Logs [1]", contentWidth),
		fitLineToWidth("", contentWidth),
	}
	contentRows := visibleHeight - len(lines)
	if contentRows < 1 {
		return model.renderRunDetailsSection(1, width, height, lines)
	}
	if len(logs) == 0 {
		lines = append(lines, fitLineToWidth("No logs available for this run.", contentWidth))
		return model.renderRunDetailsSection(1, width, height, lines)
	}
	selectedIndex := clampSelection(model.runDetailsLogsCursor, len(logs))
	start := selectedIndex - (contentRows / 2)
	if start < 0 {
		start = 0
	}
	if start > len(logs)-contentRows {
		start = len(logs) - contentRows
	}
	if start < 0 {
		start = 0
	}
	end := start + contentRows
	if end > len(logs) {
		end = len(logs)
	}
	for index := start; index < end; index++ {
		logEntry := logs[index]
		line := fmt.Sprintf("id:%s lines:%s", formatLooseDisplay(string(logEntry.ID), "?"), formatLooseDisplay(string(logEntry.LineCount), "?"))
		lines = append(lines, model.renderRunDetailsTreeRow(line, index == selectedIndex, model.runDetailsFocusSection == 1, contentWidth))
	}
	return model.renderRunDetailsSection(1, width, height, lines)
}

func (model MainLayoutModel) renderRunDetailsLogContentPanel(width int, height int) []string {
	contentWidth := width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	lines := []string{
		fitLineToWidth("Log Content [2]", contentWidth),
		fitLineToWidth("", contentWidth),
	}
	contentRows := visibleHeight - len(lines)
	if contentRows < 1 {
		return model.renderRunDetailsSection(2, width, height, lines)
	}
	if model.isLogContentLoading {
		lines = append(lines, fitLineToWidth("Loading log content...", contentWidth))
		return model.renderRunDetailsSection(2, width, height, lines)
	}
	if model.selectedLogError != "" {
		lines = append(lines, fitLineToWidth("Error: "+model.selectedLogError, contentWidth))
		return model.renderRunDetailsSection(2, width, height, lines)
	}
	if strings.TrimSpace(model.selectedLogContent) == "" {
		lines = append(lines, fitLineToWidth("Select a log and press enter to load content.", contentWidth))
		return model.renderRunDetailsSection(2, width, height, lines)
	}

	contentLines := strings.Split(model.selectedLogContent, "\n")
	maxOffset := len(contentLines) - contentRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := model.runDetailsScrollOffset
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	end := offset + contentRows
	if end > len(contentLines) {
		end = len(contentLines)
	}
	for _, contentLine := range contentLines[offset:end] {
		lines = append(lines, fitLineToWidth(sanitizeDisplayLine(contentLine), contentWidth))
	}
	return model.renderRunDetailsSection(2, width, height, lines)
}
