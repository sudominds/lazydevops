package application

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"lazydevops/internal/devops"
)

func (model MainLayoutModel) defaultProjectIndex() (int, bool) {
	defaultProjectName := strings.TrimSpace(model.mainLayoutSettings.DefaultProject)
	if defaultProjectName == "" {
		return 0, false
	}

	for index, project := range model.projects {
		if strings.EqualFold(strings.TrimSpace(project.Name), defaultProjectName) {
			return index, true
		}
	}

	return 0, false
}

func clampSelection(index int, length int) int {
	if length == 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

func runDisplayName(run devops.Run) string {
	buildNumber := strings.TrimSpace(run.BuildNumber)
	if buildNumber != "" {
		return buildNumber
	}

	name := strings.TrimSpace(run.Name)
	if name != "" {
		return name
	}

	return fmt.Sprintf("Run %d", run.ID)
}

func sortRunsMostRecent(runs []devops.Run) []devops.Run {
	if len(runs) < 2 {
		return runs
	}

	sortedRuns := make([]devops.Run, len(runs))
	copy(sortedRuns, runs)
	sort.SliceStable(sortedRuns, func(leftIndex int, rightIndex int) bool {
		leftTime, leftHasTime := mostRecentRunTime(sortedRuns[leftIndex])
		rightTime, rightHasTime := mostRecentRunTime(sortedRuns[rightIndex])
		if leftHasTime && rightHasTime {
			if !leftTime.Equal(rightTime) {
				return leftTime.After(rightTime)
			}
		} else if leftHasTime != rightHasTime {
			return leftHasTime
		}
		return sortedRuns[leftIndex].ID > sortedRuns[rightIndex].ID
	})
	return sortedRuns
}

func mostRecentRunTime(run devops.Run) (time.Time, bool) {
	candidateDates := []string{
		run.StartTime,
		run.QueueTime,
		run.CreatedDate,
		run.CreatedOn,
		run.FinishedDate,
		run.FinishTime,
	}

	for _, candidateDate := range candidateDates {
		parsedTime, ok := parseRunTime(candidateDate)
		if ok {
			return parsedTime, true
		}
	}

	return time.Time{}, false
}

func formatRunDateTime(createdDate string) string {
	dateValue := strings.TrimSpace(createdDate)
	if dateValue == "" {
		return ""
	}
	parsedTime, ok := parseRunTime(dateValue)
	if !ok {
		return dateValue
	}

	return parsedTime.Local().Format("2006-01-02 15:04:05")
}

func formatRunDuration(run devops.Run) string {
	startValue := runListStartValue(run)
	if strings.TrimSpace(startValue) == "" {
		return ""
	}
	endValue := runListFinishValue(run)
	return formatRunDurationFromBounds(startValue, endValue)
}

func runListDateValue(run devops.Run) string {
	candidateDates := []string{
		run.CreatedDate,
		run.CreatedOn,
		run.QueueTime,
		run.StartTime,
		run.FinishedDate,
		run.FinishTime,
	}

	for _, candidateDate := range candidateDates {
		trimmedDate := strings.TrimSpace(candidateDate)
		if trimmedDate != "" {
			return trimmedDate
		}
	}

	return ""
}

func runListStartValue(run devops.Run) string {
	candidateDates := []string{
		run.StartTime,
		run.QueueTime,
		run.CreatedDate,
		run.CreatedOn,
	}

	for _, candidateDate := range candidateDates {
		trimmedDate := strings.TrimSpace(candidateDate)
		if trimmedDate != "" {
			return trimmedDate
		}
	}

	return ""
}

func runListFinishValue(run devops.Run) string {
	candidateDates := []string{
		run.FinishedDate,
		run.FinishTime,
	}

	for _, candidateDate := range candidateDates {
		trimmedDate := strings.TrimSpace(candidateDate)
		if trimmedDate != "" {
			return trimmedDate
		}
	}

	return ""
}

func pipelineLatestRunStartValue(pipeline devops.Pipeline) string {
	candidateDates := []string{
		pipeline.LatestRun.StartTime,
		pipeline.LatestRun.QueueTime,
		pipeline.LatestRun.CreatedDate,
		pipeline.LatestRun.CreatedOn,
	}

	for _, candidateDate := range candidateDates {
		trimmedDate := strings.TrimSpace(candidateDate)
		if trimmedDate != "" {
			return trimmedDate
		}
	}

	return ""
}

func pipelineLatestRunFinishValue(pipeline devops.Pipeline) string {
	candidateDates := []string{
		pipeline.LatestRun.FinishedDate,
		pipeline.LatestRun.FinishTime,
	}

	for _, candidateDate := range candidateDates {
		trimmedDate := strings.TrimSpace(candidateDate)
		if trimmedDate != "" {
			return trimmedDate
		}
	}

	return ""
}

func formatPipelineLatestRunPrefix(pipeline devops.Pipeline) string {
	startValue := pipelineLatestRunStartValue(pipeline)
	finishValue := pipelineLatestRunFinishValue(pipeline)
	branchValue := formatRunBranch(pipeline.LatestRun.SourceBranch)

	prefixParts := []string{}
	formattedDate := formatRunDateTime(startValue)
	if formattedDate != "" {
		prefixParts = append(prefixParts, formattedDate)
	}

	formattedDuration := formatRunDurationFromBounds(startValue, finishValue)
	if formattedDuration != "" {
		prefixParts = append(prefixParts, formattedDuration)
	}

	if branchValue != "" {
		prefixParts = append(prefixParts, branchValue)
	}

	if len(prefixParts) == 0 {
		return ""
	}

	return fmt.Sprintf("[%s]", strings.Join(prefixParts, "] ["))
}

func parseRunTime(rawValue string) (time.Time, bool) {
	dateValue := strings.TrimSpace(rawValue)
	if dateValue == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.0000000Z",
		"2006-01-02T15:04:05Z",
	}

	for _, layout := range layouts {
		parsedTime, parseError := time.Parse(layout, dateValue)
		if parseError == nil {
			return parsedTime, true
		}
	}

	return time.Time{}, false
}

func formatCompactDuration(duration time.Duration) string {
	totalSeconds := int(duration.Round(time.Second).Seconds())
	if totalSeconds < 0 {
		totalSeconds = 0
	}

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%dm %02ds", minutes, seconds)
}

func formatRunDurationFromBounds(startRawValue string, endRawValue string) string {
	startTime, hasStart := parseRunTime(startRawValue)
	if !hasStart {
		return ""
	}

	endTime := time.Now()
	if strings.TrimSpace(endRawValue) != "" {
		parsedEndTime, hasEnd := parseRunTime(endRawValue)
		if hasEnd {
			endTime = parsedEndTime
		}
	}

	duration := endTime.Sub(startTime)
	if duration < 0 {
		return ""
	}

	return formatCompactDuration(duration)
}

func formatRunBranch(rawBranch string) string {
	branch := strings.TrimSpace(rawBranch)
	branch = strings.TrimPrefix(branch, "refs/heads/")
	return branch
}

func shortCommitHash(sourceVersion string) string {
	commitValue := strings.TrimSpace(sourceVersion)
	if commitValue == "" {
		return ""
	}
	if len(commitValue) > 8 {
		return commitValue[:8]
	}
	return commitValue
}
