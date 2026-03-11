package application

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"lazydevops/internal/devops"
)

func collectRunExecutionWarnings(timelineLoadError error, logsLoadError error) []string {
	warnings := make([]string, 0, 2)
	if timelineLoadError != nil {
		warnings = append(warnings, fmt.Sprintf("timeline unavailable: %s", timelineLoadError.Error()))
	}
	if logsLoadError != nil {
		warnings = append(warnings, fmt.Sprintf("logs unavailable: %s", logsLoadError.Error()))
	}
	return warnings
}

func summarizeExecution(timeline devops.BuildTimeline) ([]stageSummary, []jobSummary, []taskSummary) {
	if len(timeline.Records) == 0 {
		return nil, nil, nil
	}

	stages := make([]stageSummary, 0)
	jobs := make([]jobSummary, 0)
	tasks := make([]taskSummary, 0)

	for _, record := range timeline.Records {
		recordType := strings.ToLower(strings.TrimSpace(record.Type))
		order, hasOrder := parseLooseOrder(string(record.Order))
		switch recordType {
		case "stage":
			stages = append(stages, stageSummary{
				ID:       strings.TrimSpace(string(record.ID)),
				Name:     fallbackRecordName(record.Name, strings.TrimSpace(string(record.ID)), "Stage"),
				Status:   formatExecutionStatus(record.State, record.Result),
				Duration: formatRunDurationFromBounds(record.StartTime, record.FinishTime),
				Order:    order,
				HasOrder: hasOrder,
				Sequence: len(stages),
				StartAt:  strings.TrimSpace(record.StartTime),
			})
		case "job", "phase":
			jobs = append(jobs, jobSummary{
				ID:       strings.TrimSpace(string(record.ID)),
				ParentID: strings.TrimSpace(string(record.ParentID)),
				Name:     fallbackRecordName(record.Name, strings.TrimSpace(string(record.ID)), "Job"),
				Status:   formatExecutionStatus(record.State, record.Result),
				Duration: formatRunDurationFromBounds(record.StartTime, record.FinishTime),
				Attempt:  strings.TrimSpace(string(record.Attempt)),
				LogID:    strings.TrimSpace(string(record.Log.ID)),
				Order:    order,
				HasOrder: hasOrder,
				Sequence: len(jobs),
				StartAt:  strings.TrimSpace(record.StartTime),
			})
		case "task":
			tasks = append(tasks, taskSummary{
				ID:          strings.TrimSpace(string(record.ID)),
				ParentJobID: strings.TrimSpace(string(record.ParentID)),
				Name:        fallbackRecordName(record.Name, strings.TrimSpace(string(record.ID)), "Task"),
				Status:      formatExecutionStatus(record.State, record.Result),
				Duration:    formatRunDurationFromBounds(record.StartTime, record.FinishTime),
				Attempt:     strings.TrimSpace(string(record.Attempt)),
				LogID:       strings.TrimSpace(string(record.Log.ID)),
				Order:       order,
				HasOrder:    hasOrder,
				Sequence:    len(tasks),
				StartAt:     strings.TrimSpace(record.StartTime),
			})
		}
	}

	sort.SliceStable(stages, func(leftIndex int, rightIndex int) bool {
		return executionSummaryLess(
			stages[leftIndex].HasOrder,
			stages[leftIndex].Order,
			stages[leftIndex].Sequence,
			stages[leftIndex].StartAt,
			stages[leftIndex].Name,
			stages[rightIndex].HasOrder,
			stages[rightIndex].Order,
			stages[rightIndex].Sequence,
			stages[rightIndex].StartAt,
			stages[rightIndex].Name,
		)
	})
	sort.SliceStable(jobs, func(leftIndex int, rightIndex int) bool {
		return executionSummaryLess(
			jobs[leftIndex].HasOrder,
			jobs[leftIndex].Order,
			jobs[leftIndex].Sequence,
			jobs[leftIndex].StartAt,
			jobs[leftIndex].Name,
			jobs[rightIndex].HasOrder,
			jobs[rightIndex].Order,
			jobs[rightIndex].Sequence,
			jobs[rightIndex].StartAt,
			jobs[rightIndex].Name,
		)
	})
	sort.SliceStable(tasks, func(leftIndex int, rightIndex int) bool {
		return executionSummaryLess(
			tasks[leftIndex].HasOrder,
			tasks[leftIndex].Order,
			tasks[leftIndex].Sequence,
			tasks[leftIndex].StartAt,
			tasks[leftIndex].Name,
			tasks[rightIndex].HasOrder,
			tasks[rightIndex].Order,
			tasks[rightIndex].Sequence,
			tasks[rightIndex].StartAt,
			tasks[rightIndex].Name,
		)
	})

	for index := range stages {
		if stages[index].Duration == "" {
			stages[index].Duration = "N/A"
		}
	}
	for index := range jobs {
		if jobs[index].Duration == "" {
			jobs[index].Duration = "N/A"
		}
	}
	for index := range tasks {
		if tasks[index].Duration == "" {
			tasks[index].Duration = "N/A"
		}
	}

	return stages, jobs, tasks
}

func executionSummaryLess(leftHasOrder bool, leftOrder float64, leftSequence int, leftStart string, leftName string, rightHasOrder bool, rightOrder float64, rightSequence int, rightStart string, rightName string) bool {
	if leftHasOrder && rightHasOrder && !almostEqualFloat(leftOrder, rightOrder) {
		return leftOrder < rightOrder
	}
	if leftHasOrder != rightHasOrder {
		return leftHasOrder
	}
	if leftSequence != rightSequence {
		return leftSequence < rightSequence
	}
	leftStartTime, leftHasStartTime := parseRunTime(leftStart)
	rightStartTime, rightHasStartTime := parseRunTime(rightStart)
	if leftHasStartTime && rightHasStartTime && !leftStartTime.Equal(rightStartTime) {
		return leftStartTime.Before(rightStartTime)
	}
	if leftHasStartTime != rightHasStartTime {
		return leftHasStartTime
	}
	return strings.ToLower(strings.TrimSpace(leftName)) < strings.ToLower(strings.TrimSpace(rightName))
}

func almostEqualFloat(left float64, right float64) bool {
	return math.Abs(left-right) < 0.000001
}

func fallbackRecordName(name string, id string, prefix string) string {
	trimmedName := strings.TrimSpace(name)
	if trimmedName != "" {
		return trimmedName
	}
	trimmedID := strings.TrimSpace(id)
	if trimmedID != "" {
		return fmt.Sprintf("%s %s", prefix, trimmedID)
	}
	return prefix
}

func parseLooseInt(value string) (int, bool) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return 0, false
	}
	parsedValue, parseError := strconv.Atoi(trimmedValue)
	if parseError != nil {
		return 0, false
	}
	return parsedValue, true
}

func parseLooseOrder(value string) (float64, bool) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return 0, false
	}
	parsedValue, parseError := strconv.ParseFloat(trimmedValue, 64)
	if parseError != nil {
		return 0, false
	}
	return parsedValue, true
}

func formatLooseDisplay(value string, fallback string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fallback
	}
	return trimmedValue
}

func formatExecutionStatus(state string, result string) string {
	normalizedResult := strings.TrimSpace(result)
	normalizedState := strings.TrimSpace(state)

	if normalizedResult != "" && normalizedState != "" {
		return strings.ToLower(normalizedResult) + " / " + strings.ToLower(normalizedState)
	}
	if normalizedResult != "" {
		return strings.ToLower(normalizedResult)
	}
	if normalizedState != "" {
		return strings.ToLower(normalizedState)
	}
	return "unknown"
}

func buildPipelinesListCommand(organizationURL string, projectName string) string {
	return fmt.Sprintf(
		"az pipelines list --organization %s --project %s --output json",
		shellSingleQuote(organizationURL),
		shellSingleQuote(projectName),
	)
}

func buildRunsListCommand(organizationURL string, projectName string, pipelineID string, result string) string {
	commandParts := []string{
		"az pipelines runs list",
		fmt.Sprintf("--organization %s", shellSingleQuote(organizationURL)),
		fmt.Sprintf("--project %s", shellSingleQuote(projectName)),
		fmt.Sprintf("--pipeline-ids %s", pipelineID),
	}
	normalizedResult := normalizedRunResultFilter(result)
	if normalizedResult != "" {
		commandParts = append(commandParts, fmt.Sprintf("--result %s", shellSingleQuote(normalizedResult)))
	}
	commandParts = append(commandParts,
		"--query-order QueueTimeDesc",
		fmt.Sprintf("--top %d", defaultRunFetchLimit),
		"--output json",
	)
	return strings.Join(commandParts, " ")
}

func buildRunDetailsCommand(organizationURL string, projectName string, runID string) string {
	return fmt.Sprintf(
		"az pipelines build show --organization %s --project %s --id %s --output json",
		shellSingleQuote(organizationURL),
		shellSingleQuote(projectName),
		runID,
	)
}

func buildRunTimelineCommand(organizationURL string, projectName string, runID string) string {
	return fmt.Sprintf(
		"az devops invoke --organization %s --area build --resource timeline --route-parameters project=%s buildId=%s --output json",
		shellSingleQuote(organizationURL),
		shellSingleQuote(projectName),
		runID,
	)
}

func buildRunLogsListCommand(organizationURL string, projectName string, runID string) string {
	return fmt.Sprintf(
		"az devops invoke --organization %s --area build --resource logs --route-parameters project=%s buildId=%s --output json",
		shellSingleQuote(organizationURL),
		shellSingleQuote(projectName),
		runID,
	)
}

func buildRunLogCommand(organizationURL string, projectName string, runID string, logID int) string {
	return fmt.Sprintf(
		"az devops invoke --organization %s --area build --resource logs --route-parameters project=%s buildId=%s logId=%d",
		shellSingleQuote(organizationURL),
		shellSingleQuote(projectName),
		runID,
		logID,
	)
}
