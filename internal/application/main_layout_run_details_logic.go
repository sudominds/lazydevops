package application

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"lazydevops/internal/devops"
)

func (model MainLayoutModel) selectedProject() (devops.Project, bool) {
	if len(model.projects) == 0 || model.selectedProjectIndex < 0 || model.selectedProjectIndex >= len(model.projects) {
		return devops.Project{}, false
	}
	return model.projects[model.selectedProjectIndex], true
}

func (model MainLayoutModel) selectedPipeline() (devops.Pipeline, bool) {
	if len(model.pipelines) == 0 || model.selectedPipelineIndex < 0 || model.selectedPipelineIndex >= len(model.pipelines) {
		return devops.Pipeline{}, false
	}
	return model.pipelines[model.selectedPipelineIndex], true
}

func (model MainLayoutModel) selectedRun() (devops.Run, bool) {
	if len(model.runs) == 0 || model.selectedRunIndex < 0 || model.selectedRunIndex >= len(model.runs) {
		return devops.Run{}, false
	}
	return model.runs[model.selectedRunIndex], true
}

func (model *MainLayoutModel) clearRunExecutionData() {
	model.hasRunDetails = false
	model.selectedRunDetails = devops.RunDetails{}
	model.selectedRunTimeline = devops.BuildTimeline{}
	model.selectedRunLogs = nil
	model.runExecutionWarnings = nil
	model.runDetailsScrollOffset = 0
	model.runDetailsFocusSection = 0
	model.runDetailsTreeCursor = 0
	model.runDetailsLogsCursor = 0
	model.runDetailsExpandedStageIDs = map[string]bool{}
	model.runDetailsExpandedJobIDs = map[string]bool{}
	model.runDetailsLoadedRecordIDs = map[string]bool{}
	model.selectedStageKey = ""
	model.selectedJobID = ""
	model.selectedTaskID = ""
	model.selectedLogID = ""
	model.selectedLogContent = ""
	model.selectedLogError = ""
	model.isLogContentLoading = false
	model.activeTimelineRequestID = 0
	model.logContentCache = map[string]string{}
	model.highlightedLogContentCache = map[string][]string{}
	model.highlightLogFailureCache = map[string]bool{}
	model.runDetailsGoToTopPending = false
}

func (model *MainLayoutModel) initializeRunDetailsExplorer() {
	model.runDetailsFocusSection = 0
	model.runDetailsTreeCursor = 0
	model.runDetailsLogsCursor = 0
	model.selectedTaskID = ""
	model.selectedLogID = ""
	model.selectedLogContent = ""
	model.selectedLogError = ""
	model.isLogContentLoading = false
	model.runDetailsScrollOffset = 0
	model.logContentCache = map[string]string{}
	model.highlightedLogContentCache = map[string][]string{}
	model.highlightLogFailureCache = map[string]bool{}
	model.runDetailsGoToTopPending = false
	model.runDetailsExpandedStageIDs = map[string]bool{}
	model.runDetailsExpandedJobIDs = map[string]bool{}
	model.runDetailsLoadedRecordIDs = map[string]bool{}
	groups := model.buildRunDetailsStageGroups()
	if len(groups) > 0 {
		first := groups[0]
		model.selectedStageKey = first.Key
		model.runDetailsExpandedStageIDs[first.Key] = true
		if len(first.Jobs) > 0 {
			model.selectedJobID = first.Jobs[0].Job.ID
		}
	}
}

func (model *MainLayoutModel) cycleRunDetailsFocus(delta int) {
	if delta == 0 {
		return
	}
	if model.runDetailsFocusSection == 1 {
		model.runDetailsFocusSection = 2
	}
	if delta > 0 {
		if model.runDetailsFocusSection == 0 {
			model.runDetailsFocusSection = 2
		} else {
			model.runDetailsFocusSection = 0
		}
		return
	}
	if model.runDetailsFocusSection == 2 {
		model.runDetailsFocusSection = 0
	} else {
		model.runDetailsFocusSection = 2
	}
}

func (model MainLayoutModel) handleRunDetailsMove(delta int) (tea.Model, tea.Cmd) {
	switch model.runDetailsFocusSection {
	case 0:
		items := model.flattenRunDetailsTree(model.buildRunDetailsStageGroups())
		if len(items) == 0 {
			model.runDetailsTreeCursor = 0
			return model, nil
		}
		model.runDetailsTreeCursor = clampSelection(model.runDetailsTreeCursor+delta, len(items))
		selectedItem := items[model.runDetailsTreeCursor]
		if selectedItem.Kind == runDetailsTreeItemStage {
			model.selectedStageKey = selectedItem.StageKey
			model.selectedJobID = ""
			model.selectedTaskID = ""
		} else if selectedItem.Kind == runDetailsTreeItemJob {
			model.selectedJobID = selectedItem.Job.ID
			model.selectedStageKey = selectedItem.StageKey
			model.selectedTaskID = ""
		} else {
			model.selectedStageKey = selectedItem.StageKey
			model.selectedJobID = selectedItem.Job.ID
			model.selectedTaskID = selectedItem.Task.ID
		}
		model.runDetailsLogsCursor = 0
		return model, nil
	case 1:
		model.scrollRunDetails(delta)
		return model, nil
	case 2:
		model.scrollRunDetails(delta)
		return model, nil
	default:
		return model, nil
	}
}

func (model MainLayoutModel) handleRunDetailsBack() (tea.Model, tea.Cmd) {
	switch model.runDetailsFocusSection {
	case 0:
		items := model.flattenRunDetailsTree(model.buildRunDetailsStageGroups())
		if len(items) == 0 {
			model.cycleRunDetailsFocus(-1)
			return model, nil
		}
		selected := items[clampSelection(model.runDetailsTreeCursor, len(items))]
		if selected.Kind == runDetailsTreeItemStage {
			if model.runDetailsExpandedStageIDs != nil && model.runDetailsExpandedStageIDs[selected.StageKey] {
				model.runDetailsExpandedStageIDs[selected.StageKey] = false
				return model, nil
			}
		}
		if selected.Kind == runDetailsTreeItemJob {
			jobKey := model.runDetailsJobKey(selected.StageKey, selected.Job.ID)
			if model.runDetailsExpandedJobIDs != nil && model.runDetailsExpandedJobIDs[jobKey] {
				model.runDetailsExpandedJobIDs[jobKey] = false
				return model, nil
			}
		}
		if model.runDetailsTreeCursor > 0 {
			model.runDetailsTreeCursor = clampSelection(model.runDetailsTreeCursor-1, len(items))
			model.syncRunDetailsSelectionFromCursor(items)
			return model, nil
		}
		model.cycleRunDetailsFocus(-1)
		return model, nil
	case 1:
		model.runDetailsFocusSection = 0
		return model, nil
	case 2:
		model.runDetailsFocusSection = 0
		return model, nil
	default:
		model.cycleRunDetailsFocus(-1)
		return model, nil
	}
}

func (model MainLayoutModel) handleRunDetailsEnter() (tea.Model, tea.Cmd) {
	switch model.runDetailsFocusSection {
	case 0:
		items := model.flattenRunDetailsTree(model.buildRunDetailsStageGroups())
		if len(items) == 0 {
			return model, nil
		}
		selected := items[clampSelection(model.runDetailsTreeCursor, len(items))]
		if selected.Kind == runDetailsTreeItemStage {
			if model.runDetailsExpandedStageIDs == nil {
				model.runDetailsExpandedStageIDs = map[string]bool{}
			}
			nextExpandedState := !model.runDetailsExpandedStageIDs[selected.StageKey]
			model.runDetailsExpandedStageIDs[selected.StageKey] = nextExpandedState
			model.selectedStageKey = selected.StageKey
			model.selectedJobID = ""
			model.selectedTaskID = ""
			if nextExpandedState {
				return model.loadRunDetailsNodeTimeline(selected.Stage.ID)
			}
			return model, nil
		}
		if selected.Kind == runDetailsTreeItemJob {
			if model.runDetailsExpandedJobIDs == nil {
				model.runDetailsExpandedJobIDs = map[string]bool{}
			}
			jobKey := model.runDetailsJobKey(selected.StageKey, selected.Job.ID)
			nextExpandedState := !model.isRunDetailsJobExpanded(selected.StageKey, selected.Job.ID)
			model.runDetailsExpandedJobIDs[jobKey] = nextExpandedState
			model.selectedStageKey = selected.StageKey
			model.selectedJobID = selected.Job.ID
			model.selectedTaskID = ""
			model.runDetailsLogsCursor = 0
			if nextExpandedState {
				return model.loadRunDetailsNodeTimeline(selected.Job.ID)
			}
			return model, nil
		}
		model.selectedStageKey = selected.StageKey
		model.selectedJobID = selected.Job.ID
		model.selectedTaskID = selected.Task.ID
		model.runDetailsLogsCursor = 0
		if strings.TrimSpace(selected.Task.LogID) != "" {
			return model.loadSelectedRunLog(selected.Task.LogID)
		}
		model.runDetailsFocusSection = 1
		return model, nil
	case 1:
		return model, nil
	case 2:
		return model, nil
	default:
		return model, nil
	}
}

func (model *MainLayoutModel) syncRunDetailsSelectionFromCursor(items []runDetailsTreeItem) {
	if len(items) == 0 {
		model.selectedStageKey = ""
		model.selectedJobID = ""
		model.selectedTaskID = ""
		return
	}
	selectedItem := items[clampSelection(model.runDetailsTreeCursor, len(items))]
	if selectedItem.Kind == runDetailsTreeItemStage {
		model.selectedStageKey = selectedItem.StageKey
		model.selectedJobID = ""
		model.selectedTaskID = ""
		return
	}
	if selectedItem.Kind == runDetailsTreeItemJob {
		model.selectedJobID = selectedItem.Job.ID
		model.selectedStageKey = selectedItem.StageKey
		model.selectedTaskID = ""
		return
	}
	model.selectedStageKey = selectedItem.StageKey
	model.selectedJobID = selectedItem.Job.ID
	model.selectedTaskID = selectedItem.Task.ID
}

func (model MainLayoutModel) loadRunDetailsNodeTimeline(parentRecordID string) (tea.Model, tea.Cmd) {
	trimmedParentRecordID := strings.TrimSpace(parentRecordID)
	if trimmedParentRecordID == "" {
		return model, nil
	}
	if model.runDetailsLoadedRecordIDs != nil && model.runDetailsLoadedRecordIDs[trimmedParentRecordID] {
		return model, nil
	}
	selectedProject, hasProject := model.selectedProject()
	selectedRun, hasRun := model.selectedRun()
	if !hasProject || !hasRun {
		return model, nil
	}
	parentRecord, hasParentRecord := findTimelineRecordByID(model.selectedRunTimeline, trimmedParentRecordID)
	if !hasParentRecord {
		return model, nil
	}
	detailsTimelineID := strings.TrimSpace(string(parentRecord.Details.ID))
	detailsURL := strings.TrimSpace(parentRecord.Details.URL)
	if detailsTimelineID == "" && detailsURL == "" {
		if model.runDetailsLoadedRecordIDs == nil {
			model.runDetailsLoadedRecordIDs = map[string]bool{}
		}
		model.runDetailsLoadedRecordIDs[trimmedParentRecordID] = true
		return model, nil
	}

	model.activeTimelineRequestID++
	requestID := model.activeTimelineRequestID
	model.isLoading = true
	model.loadingMessage = "Loading execution details..."
	return model, model.loadRunDetailsNodeTimelineCommand(requestID, selectedProject.Name, selectedRun.ID, trimmedParentRecordID, detailsTimelineID, detailsURL)
}

func (model MainLayoutModel) loadRunDetailsNodeTimelineCommand(requestID int, projectName string, buildID int, parentRecordID string, timelineID string, detailsURL string) tea.Cmd {
	return func() tea.Msg {
		timeline, loadError := model.devopsService.GetBuildTimelineDetails(projectName, buildID, timelineID, detailsURL)
		return runDetailsTimelineExpandedMessage{
			requestID:      requestID,
			parentRecordID: parentRecordID,
			timeline:       timeline,
			loadError:      loadError,
		}
	}
}

func (model MainLayoutModel) loadSelectedRunLog(logID string) (tea.Model, tea.Cmd) {
	trimmedLogID := strings.TrimSpace(logID)
	if trimmedLogID == "" {
		model.selectedLogError = "selected log id is empty"
		return model, nil
	}
	if cachedContent, exists := model.logContentCache[trimmedLogID]; exists {
		model.selectedLogID = trimmedLogID
		model.selectedLogContent = cachedContent
		model.selectedLogError = ""
		model.runDetailsFocusSection = 2
		model.runDetailsScrollOffset = 0
		return model, nil
	}

	model.activeLogRequestID++
	requestID := model.activeLogRequestID
	model.isLogContentLoading = true
	model.selectedLogError = ""
	model.selectedLogID = trimmedLogID
	model.runDetailsFocusSection = 2
	model.runDetailsScrollOffset = 0
	return model, model.loadRunLogContentCommand(requestID, trimmedLogID)
}

func (model MainLayoutModel) loadRunLogContentCommand(requestID int, logID string) tea.Cmd {
	selectedProject, hasProject := model.selectedProject()
	selectedRun, hasRun := model.selectedRun()
	if !hasProject || !hasRun {
		return func() tea.Msg {
			return runLogLoadedMessage{requestID: requestID, logID: logID, loadError: fmt.Errorf("project or run not selected")}
		}
	}
	numericLogID, hasNumericLogID := parseLooseInt(logID)
	if !hasNumericLogID {
		return func() tea.Msg {
			return runLogLoadedMessage{requestID: requestID, logID: logID, loadError: fmt.Errorf("log id is non-numeric: %s", logID)}
		}
	}

	return func() tea.Msg {
		content, loadError := model.devopsService.GetBuildLog(selectedProject.Name, selectedRun.ID, numericLogID)
		return runLogLoadedMessage{
			requestID: requestID,
			logID:     logID,
			content:   content,
			loadError: loadError,
		}
	}
}
