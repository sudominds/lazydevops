package application

import (
	"fmt"
	"sort"
	"strings"

	"lazydevops/internal/devops"
)

func (model MainLayoutModel) buildRunDetailsStageGroups() []runDetailsStageGroup {
	stages, jobs, tasks := summarizeExecution(model.selectedRunTimeline)
	if len(stages) == 0 {
		return nil
	}

	parentByRecordID, typeByRecordID := buildTimelineParentAndTypeMaps(model.selectedRunTimeline)
	stageIDSet := make(map[string]bool, len(stages))
	for _, stage := range stages {
		stageID := strings.TrimSpace(stage.ID)
		if stageID != "" {
			stageIDSet[stageID] = true
		}
	}
	jobIDSet := make(map[string]bool, len(jobs))
	for _, job := range jobs {
		jobID := strings.TrimSpace(job.ID)
		if jobID != "" {
			jobIDSet[jobID] = true
		}
	}

	groups := make([]runDetailsStageGroup, 0, len(stages))
	for stageIndex, stage := range stages {
		stageKey := strings.TrimSpace(stage.ID)
		if stageKey == "" {
			stageKey = fmt.Sprintf("stage-%d", stageIndex)
		}
		group := runDetailsStageGroup{
			Key:   stageKey,
			Stage: stage,
			Jobs:  make([]runDetailsJobGroup, 0),
		}
		displayedJobs := make([]jobSummary, 0)
		for _, job := range jobs {
			jobStageID := findAncestorIDByTypeOrSet(strings.TrimSpace(job.ID), parentByRecordID, typeByRecordID, "stage", stageIDSet)
			if jobStageID == "" {
				jobStageID = strings.TrimSpace(job.ParentID)
			}
			if strings.TrimSpace(jobStageID) != strings.TrimSpace(stage.ID) {
				continue
			}
			// Keep only the first job/phase layer under a stage to match Azure job rows.
			if hasAncestorWithAnyTypeBeforeStage(strings.TrimSpace(job.ID), strings.TrimSpace(stage.ID), parentByRecordID, typeByRecordID, map[string]bool{"job": true, "phase": true}) {
				continue
			}
			displayedJobs = append(displayedJobs, job)
		}

		displayedJobIDSet := make(map[string]bool, len(displayedJobs))
		for _, job := range displayedJobs {
			jobID := strings.TrimSpace(job.ID)
			if jobID != "" {
				displayedJobIDSet[jobID] = true
			}
		}
		tasksByDisplayedJobID := make(map[string][]taskSummary)
		for _, task := range tasks {
			displayJobID := findAncestorIDByTypeOrSet(strings.TrimSpace(task.ID), parentByRecordID, typeByRecordID, "", displayedJobIDSet)
			if displayJobID == "" {
				parentJobID := findAncestorIDByTypeOrSet(strings.TrimSpace(task.ID), parentByRecordID, typeByRecordID, "job", jobIDSet)
				if strings.TrimSpace(parentJobID) != "" && displayedJobIDSet[strings.TrimSpace(parentJobID)] {
					displayJobID = strings.TrimSpace(parentJobID)
				}
			}
			if displayJobID == "" {
				continue
			}
			tasksByDisplayedJobID[displayJobID] = append(tasksByDisplayedJobID[displayJobID], task)
		}
		for _, job := range displayedJobs {
			jobGroup := runDetailsJobGroup{
				Job:   job,
				Tasks: append([]taskSummary{}, tasksByDisplayedJobID[strings.TrimSpace(job.ID)]...),
			}
			sort.SliceStable(jobGroup.Tasks, func(leftIndex int, rightIndex int) bool {
				return taskSummaryLess(jobGroup.Tasks[leftIndex], jobGroup.Tasks[rightIndex])
			})
			group.Jobs = append(group.Jobs, jobGroup)
		}
		sort.SliceStable(group.Jobs, func(leftIndex int, rightIndex int) bool {
			return executionSummaryLess(
				group.Jobs[leftIndex].Job.HasOrder,
				group.Jobs[leftIndex].Job.Order,
				group.Jobs[leftIndex].Job.Sequence,
				group.Jobs[leftIndex].Job.StartAt,
				group.Jobs[leftIndex].Job.Name,
				group.Jobs[rightIndex].Job.HasOrder,
				group.Jobs[rightIndex].Job.Order,
				group.Jobs[rightIndex].Job.Sequence,
				group.Jobs[rightIndex].Job.StartAt,
				group.Jobs[rightIndex].Job.Name,
			)
		})
		groups = append(groups, group)
	}
	return groups
}

func buildTimelineParentAndTypeMaps(timeline devops.BuildTimeline) (map[string]string, map[string]string) {
	parentByRecordID := map[string]string{}
	typeByRecordID := map[string]string{}
	for _, record := range timeline.Records {
		recordID := strings.TrimSpace(string(record.ID))
		if recordID == "" {
			continue
		}
		parentByRecordID[recordID] = strings.TrimSpace(string(record.ParentID))
		typeByRecordID[recordID] = strings.ToLower(strings.TrimSpace(record.Type))
	}
	return parentByRecordID, typeByRecordID
}

func findAncestorIDByTypeOrSet(startID string, parentByRecordID map[string]string, typeByRecordID map[string]string, targetType string, allowedIDs map[string]bool) string {
	visited := map[string]bool{}
	currentID := strings.TrimSpace(startID)
	normalizedTargetType := strings.ToLower(strings.TrimSpace(targetType))
	for currentID != "" {
		if visited[currentID] {
			break
		}
		visited[currentID] = true
		if normalizedTargetType != "" && strings.TrimSpace(typeByRecordID[currentID]) == normalizedTargetType {
			return currentID
		}
		if len(allowedIDs) > 0 && allowedIDs[currentID] {
			return currentID
		}
		currentID = strings.TrimSpace(parentByRecordID[currentID])
	}
	return ""
}

func hasAncestorWithAnyTypeBeforeStage(startID string, stageID string, parentByRecordID map[string]string, typeByRecordID map[string]string, targetTypes map[string]bool) bool {
	if len(targetTypes) == 0 {
		return false
	}
	trimmedStartID := strings.TrimSpace(startID)
	trimmedStageID := strings.TrimSpace(stageID)
	if trimmedStartID == "" || trimmedStageID == "" {
		return false
	}
	visited := map[string]bool{}
	currentID := strings.TrimSpace(parentByRecordID[trimmedStartID])
	for currentID != "" {
		if visited[currentID] {
			break
		}
		visited[currentID] = true
		if strings.TrimSpace(currentID) == trimmedStageID {
			return false
		}
		currentType := strings.ToLower(strings.TrimSpace(typeByRecordID[currentID]))
		if targetTypes[currentType] {
			return true
		}
		currentID = strings.TrimSpace(parentByRecordID[currentID])
	}
	return false
}

func taskSummaryLess(left taskSummary, right taskSummary) bool {
	leftStartTime, leftHasStartTime := parseRunTime(left.StartAt)
	rightStartTime, rightHasStartTime := parseRunTime(right.StartAt)
	if leftHasStartTime && rightHasStartTime && !leftStartTime.Equal(rightStartTime) {
		return leftStartTime.Before(rightStartTime)
	}
	if leftHasStartTime != rightHasStartTime {
		return leftHasStartTime
	}
	return executionSummaryLess(
		left.HasOrder,
		left.Order,
		left.Sequence,
		left.StartAt,
		left.Name,
		right.HasOrder,
		right.Order,
		right.Sequence,
		right.StartAt,
		right.Name,
	)
}

func (model MainLayoutModel) flattenRunDetailsTree(groups []runDetailsStageGroup) []runDetailsTreeItem {
	items := make([]runDetailsTreeItem, 0)
	for _, group := range groups {
		items = append(items, runDetailsTreeItem{
			Kind:     runDetailsTreeItemStage,
			StageKey: group.Key,
			Stage:    group.Stage,
		})
		if !model.runDetailsExpandedStageIDs[group.Key] {
			continue
		}
		for _, jobGroup := range group.Jobs {
			items = append(items, runDetailsTreeItem{
				Kind:     runDetailsTreeItemJob,
				StageKey: group.Key,
				Stage:    group.Stage,
				Job:      jobGroup.Job,
			})
			if !model.isRunDetailsJobExpanded(group.Key, jobGroup.Job.ID) {
				continue
			}
			for _, task := range jobGroup.Tasks {
				items = append(items, runDetailsTreeItem{
					Kind:     runDetailsTreeItemTask,
					StageKey: group.Key,
					Stage:    group.Stage,
					Job:      jobGroup.Job,
					Task:     task,
				})
			}
		}
	}
	return items
}

func (model MainLayoutModel) currentRunDetailsLogs() []devops.BuildLog {
	if len(model.selectedRunLogs) == 0 {
		return nil
	}
	return model.selectedRunLogs
}

func (model MainLayoutModel) runDetailsJobKey(stageKey string, jobID string) string {
	return strings.TrimSpace(stageKey) + "|" + strings.TrimSpace(jobID)
}

func (model MainLayoutModel) isRunDetailsJobExpanded(stageKey string, jobID string) bool {
	if model.runDetailsExpandedJobIDs == nil {
		return false
	}
	jobKey := model.runDetailsJobKey(stageKey, jobID)
	expanded, hasExplicitSetting := model.runDetailsExpandedJobIDs[jobKey]
	if hasExplicitSetting {
		return expanded
	}
	return false
}

func findTimelineRecordByID(timeline devops.BuildTimeline, recordID string) (devops.TimelineRecord, bool) {
	trimmedRecordID := strings.TrimSpace(recordID)
	if trimmedRecordID == "" {
		return devops.TimelineRecord{}, false
	}
	for _, record := range timeline.Records {
		if strings.TrimSpace(string(record.ID)) == trimmedRecordID {
			return record, true
		}
	}
	return devops.TimelineRecord{}, false
}

func mergeBuildTimelineRecords(base devops.BuildTimeline, overlay devops.BuildTimeline) devops.BuildTimeline {
	if len(overlay.Records) == 0 {
		return base
	}
	if strings.TrimSpace(base.ID) == "" {
		base.ID = overlay.ID
	}
	merged := make([]devops.TimelineRecord, 0, len(base.Records)+len(overlay.Records))
	merged = append(merged, base.Records...)
	seenRecordIDs := map[string]bool{}
	for _, record := range base.Records {
		recordID := strings.TrimSpace(string(record.ID))
		if recordID != "" {
			seenRecordIDs[recordID] = true
		}
	}
	for _, record := range overlay.Records {
		recordID := strings.TrimSpace(string(record.ID))
		if recordID != "" && seenRecordIDs[recordID] {
			continue
		}
		merged = append(merged, record)
		if recordID != "" {
			seenRecordIDs[recordID] = true
		}
	}
	base.Records = merged
	return base
}

func (model *MainLayoutModel) scrollRunDetails(delta int) {
	if delta == 0 {
		return
	}
	maxOffset := model.runDetailsLogMaxOffset()
	model.runDetailsScrollOffset += delta
	if model.runDetailsScrollOffset < 0 {
		model.runDetailsScrollOffset = 0
	}
	if model.runDetailsScrollOffset > maxOffset {
		model.runDetailsScrollOffset = maxOffset
	}
}

func (model MainLayoutModel) runDetailsLogViewportLineCount() int {
	contentLineBudget := model.runDetailsContentBudget() / 2
	if contentLineBudget < 1 {
		contentLineBudget = 1
	}
	return contentLineBudget
}

func (model MainLayoutModel) runDetailsLogMaxOffset() int {
	lineCount := len(strings.Split(model.selectedLogContent, "\n"))
	maxOffset := lineCount - model.runDetailsLogViewportLineCount()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (model *MainLayoutModel) scrollRunDetailsHalfPage(direction int) {
	if direction == 0 {
		return
	}
	delta := model.runDetailsLogViewportLineCount() / 2
	if delta < 1 {
		delta = 1
	}
	model.scrollRunDetails(direction * delta)
}

func (model *MainLayoutModel) scrollRunDetailsToTop() {
	model.runDetailsScrollOffset = 0
}

func (model *MainLayoutModel) scrollRunDetailsToBottom() {
	model.runDetailsScrollOffset = model.runDetailsLogMaxOffset()
}

func (model MainLayoutModel) runDetailsContentBudget() int {
	footerLineCount := 4
	if model.isCommandPaletteVisible {
		footerLineCount += 6
	}
	panelBorderLineCount := 2
	staticHeaderLines := 4 // title + spacer + search + spacer
	panelHeight := model.windowHeight - staticHeaderLines - footerLineCount - panelBorderLineCount
	if panelHeight < 10 {
		panelHeight = 10
	}
	visibleLineCount := panelHeight - 2
	if visibleLineCount < 6 {
		visibleLineCount = 6
	}
	contentLineBudget := visibleLineCount - 5
	if contentLineBudget < 0 {
		contentLineBudget = 0
	}
	return contentLineBudget
}
