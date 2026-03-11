package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (model MainLayoutModel) renderProjectLines(contentWidth int, contentLineBudget int) []string {
	filteredIndexes := model.filteredItemIndexes()
	if len(filteredIndexes) == 0 {
		return []string{"No matching projects for current search."}
	}

	cards := make([]listCard, 0, len(filteredIndexes))
	for _, absoluteIndex := range filteredIndexes {
		project := model.projects[absoluteIndex]
		cards = append(cards, listCard{
			Lines: []listCardLine{
				{Text: project.Name},
			},
		})
	}

	return model.renderVisibleCardLines(cards, contentWidth, contentLineBudget, model.searchInput.Value())
}

func (model MainLayoutModel) renderPipelineLines(contentWidth int, contentLineBudget int) []string {
	filteredIndexes := model.filteredItemIndexes()
	if len(filteredIndexes) == 0 {
		if model.isLoading {
			return []string{}
		}
		return []string{"No matching pipelines for current search."}
	}

	cards := make([]listCard, 0, len(filteredIndexes))
	for _, absoluteIndex := range filteredIndexes {
		pipeline := model.pipelines[absoluteIndex]
		meta := fmt.Sprintf("Pipeline ID: %d", pipeline.ID)
		cards = append(cards, listCard{
			Lines: []listCardLine{
				{Text: pipeline.Name},
				{Text: meta, Muted: true},
			},
		})
	}

	return model.renderVisibleCardLines(cards, contentWidth, contentLineBudget, model.searchInput.Value())
}

func (model MainLayoutModel) renderRunLines(contentWidth int, contentLineBudget int) []string {
	filteredIndexes := model.filteredItemIndexes()
	if len(filteredIndexes) == 0 {
		return []string{"No matching runs for current search."}
	}

	cards := make([]listCard, 0, len(filteredIndexes))
	for _, absoluteIndex := range filteredIndexes {
		run := model.runs[absoluteIndex]
		name := runDisplayName(run)
		status := strings.TrimSpace(run.Result)
		if status == "" {
			status = strings.TrimSpace(run.State)
		}
		if status == "" {
			status = strings.TrimSpace(run.Status)
		}
		runDateTime := formatRunDateTime(runListStartValue(run))
		runDuration := formatRunDuration(run)
		branch := formatRunBranch(run.SourceBranch)
		reason := strings.TrimSpace(run.Reason)
		commit := shortCommitHash(run.SourceVersion)

		prefixParts := []string{}
		if runDateTime != "" {
			prefixParts = append(prefixParts, runDateTime)
		}
		if runDuration != "" {
			prefixParts = append(prefixParts, runDuration)
		}

		nameWithStatus := model.statusPrefix(status, name)
		stagesLine := model.runCardStagesLine(run.ID)
		cards = append(cards, listCard{
			Lines: []listCardLine{
				{Text: nameWithStatus},
				{Text: strings.Join(prefixParts, " | "), Muted: true},
				{Text: fmt.Sprintf("RunId %d | Trigger: %s", run.ID, configuredValue(reason)), Muted: true},
				{Text: fmt.Sprintf("branch: %s | commit: %s", configuredValue(branch), configuredValue(commit)), Muted: true},
				{Text: stagesLine},
			},
		})
	}

	filter := parseRunSearchFilter(model.searchInput.Value())
	return model.renderVisibleCardLines(cards, contentWidth, contentLineBudget, filter.TextQuery)
}

func (model MainLayoutModel) runCardStagesLine(runID int) string {
	if model.runStagePreviews == nil {
		return "Stages: loading..."
	}
	preview, hasPreview := model.runStagePreviews[runID]
	if !hasPreview || preview.IsLoading {
		return "Stages: loading..."
	}
	if strings.TrimSpace(preview.LoadError) != "" {
		return "Stages: unavailable"
	}
	if len(preview.Statuses) == 0 {
		return "Stages: N/A"
	}

	stageIcons := make([]string, 0, len(preview.Statuses))
	for _, stageStatus := range preview.Statuses {
		stageIcons = append(stageIcons, model.statusIcon(stageStatus))
	}
	return "Stages: " + strings.Join(stageIcons, " - ")
}

func (model MainLayoutModel) renderVisibleCardLines(cards []listCard, contentWidth int, contentLineBudget int, searchQuery string) []string {
	if contentLineBudget < 1 {
		return []string{}
	}
	if len(cards) == 0 {
		return []string{}
	}

	hasSelectedCard := model.currentMode == inputModeNormal && model.currentStage != stageRunDetails
	selectedCardIndex := clampSelection(model.listCursorIndex, len(cards))
	if !hasSelectedCard {
		selectedCardIndex = -1
	}
	allLines := make([]string, 0, len(cards)*4)
	cardSpans := make([]lineSpan, len(cards))
	topSeparatorColor := model.palette.Muted
	if hasSelectedCard && selectedCardIndex == 0 {
		if model.listRainbowAccentEnabled() {
			topSeparatorColor = model.listRainbowColor(0)
		} else {
			topSeparatorColor = model.palette.Accent
		}
	}
	topSeparatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(topSeparatorColor))
	allLines = append(allLines, topSeparatorStyle.Render(strings.Repeat("─", contentWidth)))
	for cardIndex, card := range cards {
		startLine := len(allLines)
		cardLines := model.renderSelectableCardLines(contentWidth, hasSelectedCard && cardIndex == selectedCardIndex, card.Lines, searchQuery, cardIndex)
		allLines = append(allLines, cardLines...)
		cardSpans[cardIndex] = lineSpan{Start: startLine, End: len(allLines) - 1}
		separatorColor := model.palette.Muted
		nextCardIndex := cardIndex + 1
		if hasSelectedCard && (cardIndex == selectedCardIndex || nextCardIndex == selectedCardIndex) {
			if model.listRainbowAccentEnabled() {
				separatorColor = model.listRainbowColor(cardIndex)
			} else {
				separatorColor = model.palette.Accent
			}
		}
		separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(separatorColor))
		separatorLine := separatorStyle.Render(strings.Repeat("─", contentWidth))
		allLines = append(allLines, separatorLine)
	}

	startLine := 0
	if hasSelectedCard {
		selectedSpan := cardSpans[selectedCardIndex]
		startLine = selectedSpan.End - contentLineBudget + 1
		if selectedSpan.Start < startLine {
			startLine = selectedSpan.Start
		}
	}
	if startLine < 0 {
		startLine = 0
	}
	maxOffset := len(allLines) - contentLineBudget
	if maxOffset < 0 {
		maxOffset = 0
	}
	if startLine > maxOffset {
		startLine = maxOffset
	}

	endLine := startLine + contentLineBudget
	if endLine > len(allLines) {
		endLine = len(allLines)
	}

	visibleLines := make([]string, 0, contentLineBudget)
	visibleLines = append(visibleLines, allLines[startLine:endLine]...)
	for len(visibleLines) < contentLineBudget {
		visibleLines = append(visibleLines, "")
	}

	return visibleLines
}

func (model MainLayoutModel) renderRunDetailsLines() []string {
	if !model.hasRunDetails {
		return []string{"No run details loaded.", "Press esc to go back to runs."}
	}

	runDetails := model.selectedRunDetails
	pipelineName := strings.TrimSpace(runDetails.Pipeline.Name)
	pipelineID := runDetails.Pipeline.ID
	if pipelineName == "" {
		pipelineName = strings.TrimSpace(runDetails.Definition.Name)
		pipelineID = runDetails.Definition.ID
	}
	pipelineLabel := "N/A"
	if pipelineName != "" {
		pipelineLabel = fmt.Sprintf("%s [ID:%d]", pipelineName, pipelineID)
	}
	requestedBy := strings.TrimSpace(runDetails.RequestedBy.DisplayName)
	if requestedBy == "" {
		requestedBy = strings.TrimSpace(runDetails.RequestedBy.UniqueName)
	}
	if requestedBy == "" {
		requestedBy = "N/A"
	}
	commitValue := strings.TrimSpace(runDetails.SourceVersion)
	if len(commitValue) > 8 {
		commitValue = commitValue[:8]
	}
	if commitValue == "" {
		commitValue = "N/A"
	}
	branch := formatRunBranch(runDetails.SourceBranch)
	if branch == "" {
		branch = "N/A"
	}
	reason := strings.TrimSpace(runDetails.Reason)
	if reason == "" {
		reason = "N/A"
	}
	state := configuredValue(runDetails.State)
	result := configuredValue(runDetails.Result)
	status := strings.TrimSpace(runDetails.Result)
	if status == "" {
		status = strings.TrimSpace(runDetails.State)
	}
	status = model.statusDisplay(status)
	queuedAt := formatRunDateTime(firstNonEmpty(runDetails.QueueTime, runDetails.CreatedDate))
	if queuedAt == "" {
		queuedAt = "N/A"
	}
	startedAtRaw := firstNonEmpty(runDetails.StartTime, runDetails.QueueTime, runDetails.CreatedDate)
	startedAt := formatRunDateTime(startedAtRaw)
	if startedAt == "" {
		startedAt = "N/A"
	}
	finishedAtRaw := strings.TrimSpace(runDetails.FinishedAt)
	finishedAt := formatRunDateTime(finishedAtRaw)
	if finishedAt == "" {
		finishedAt = "N/A"
	}
	duration := formatRunDurationFromBounds(startedAtRaw, finishedAtRaw)
	if duration == "" {
		duration = "N/A"
	}

	lines := []string{
		fmt.Sprintf("Pipeline: %s", pipelineLabel),
		fmt.Sprintf("Name: %s", runDetails.Name),
		fmt.Sprintf("Run ID: %d", runDetails.ID),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("State: %s", state),
		fmt.Sprintf("Result: %s", result),
		fmt.Sprintf("Trigger Reason: %s", reason),
		fmt.Sprintf("Branch: %s", branch),
		fmt.Sprintf("Commit: %s", commitValue),
		fmt.Sprintf("Requested By: %s", requestedBy),
		fmt.Sprintf("Queued: %s", queuedAt),
		fmt.Sprintf("Started: %s", startedAt),
		fmt.Sprintf("Finished: %s", finishedAt),
		fmt.Sprintf("Duration: %s", duration),
	}

	stages, jobs, tasks := summarizeExecution(model.selectedRunTimeline)
	lines = append(lines, "", "Stages:")
	if len(stages) == 0 {
		lines = append(lines, "None available.")
	} else {
		for _, stage := range stages {
			lines = append(lines, fmt.Sprintf("- %s [%s] (%s)", stage.Name, stage.Status, stage.Duration))
		}
	}

	lines = append(lines, "", "Jobs:")
	if len(jobs) == 0 {
		lines = append(lines, "None available.")
	} else {
		maxJobs := len(jobs)
		if maxJobs > 8 {
			maxJobs = 8
		}
		for index := 0; index < maxJobs; index++ {
			job := jobs[index]
			logSuffix := ""
			if strings.TrimSpace(job.LogID) != "" {
				logSuffix = fmt.Sprintf(" | log:%s", job.LogID)
			}
			attemptSuffix := ""
			if strings.TrimSpace(job.Attempt) != "" {
				attemptSuffix = fmt.Sprintf(" | attempt:%s", job.Attempt)
			}
			lines = append(lines, fmt.Sprintf("- %s [%s] (%s)%s%s", job.Name, job.Status, job.Duration, attemptSuffix, logSuffix))
		}
		if len(jobs) > maxJobs {
			lines = append(lines, fmt.Sprintf("... %d more jobs", len(jobs)-maxJobs))
		}
	}

	lines = append(lines, "", "Tasks:")
	if len(tasks) == 0 {
		lines = append(lines, "None available.")
	} else {
		maxTasks := len(tasks)
		if maxTasks > 8 {
			maxTasks = 8
		}
		for index := 0; index < maxTasks; index++ {
			task := tasks[index]
			logSuffix := ""
			if strings.TrimSpace(task.LogID) != "" {
				logSuffix = fmt.Sprintf(" | log:%s", task.LogID)
			}
			attemptSuffix := ""
			if strings.TrimSpace(task.Attempt) != "" {
				attemptSuffix = fmt.Sprintf(" | attempt:%s", task.Attempt)
			}
			lines = append(lines, fmt.Sprintf("- %s [%s] (%s)%s%s", task.Name, task.Status, task.Duration, attemptSuffix, logSuffix))
		}
		if len(tasks) > maxTasks {
			lines = append(lines, fmt.Sprintf("... %d more tasks", len(tasks)-maxTasks))
		}
	}

	lines = append(lines, "", "Logs:")
	if len(model.selectedRunLogs) == 0 {
		lines = append(lines, "None available.")
	} else {
		maxLogs := len(model.selectedRunLogs)
		if maxLogs > 6 {
			maxLogs = 6
		}
		for index := 0; index < maxLogs; index++ {
			log := model.selectedRunLogs[index]
			lines = append(lines, fmt.Sprintf("- id:%s lines:%s", formatLooseDisplay(string(log.ID), "unknown"), formatLooseDisplay(string(log.LineCount), "unknown")))
		}
		if len(model.selectedRunLogs) > maxLogs {
			lines = append(lines, fmt.Sprintf("... %d more logs", len(model.selectedRunLogs)-maxLogs))
		}
	}

	if len(model.runExecutionWarnings) > 0 {
		lines = append(lines, "", "Execution Data Warnings:")
		for _, warning := range model.runExecutionWarnings {
			lines = append(lines, "- "+warning)
		}
	}

	projectName, runID := model.runDetailsCommandContext()
	lines = append(lines,
		"",
		"CLI:",
		buildRunTimelineCommand(model.organizationURL, projectName, runID),
		buildRunLogsListCommand(model.organizationURL, projectName, runID),
	)
	if len(model.selectedRunLogs) > 0 {
		logID, hasNumericLogID := parseLooseInt(string(model.selectedRunLogs[0].ID))
		if hasNumericLogID {
			lines = append(lines, buildRunLogCommand(model.organizationURL, projectName, runID, logID))
		}
	}

	lines = append(lines, "", "Web URL:", runDetails.WebURL)
	return lines
}
