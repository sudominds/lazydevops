package application

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"lazydevops/internal/devops"
)

func (model MainLayoutModel) runDetailsCommandContext() (string, string) {
	selectedProject, hasProject := model.selectedProject()
	selectedRun, hasRun := model.selectedRun()
	projectName := "<project>"
	runID := "<run-id>"
	if hasProject {
		projectName = selectedProject.Name
	}
	if hasRun {
		runID = fmt.Sprintf("%d", selectedRun.ID)
	}
	return projectName, runID
}

func (model MainLayoutModel) loadProjectsCommand(requestID int) tea.Cmd {
	return func() tea.Msg {
		projects, loadError := model.devopsService.ListProjects()
		return projectsLoadedMessage{
			requestID: requestID,
			projects:  projects,
			loadError: loadError,
		}
	}
}

func (model MainLayoutModel) loadPipelinesCommand(requestID int) tea.Cmd {
	selectedProject, hasProject := model.selectedProject()
	if !hasProject {
		return func() tea.Msg {
			return pipelinesLoadedMessage{requestID: requestID, pipelines: nil, loadError: nil}
		}
	}

	return func() tea.Msg {
		pipelines, loadError := model.devopsService.ListPipelines(selectedProject.Name)
		return pipelinesLoadedMessage{
			requestID: requestID,
			pipelines: pipelines,
			loadError: loadError,
		}
	}
}

func (model MainLayoutModel) loadRunsCommand(requestID int) tea.Cmd {
	selectedProject, hasProject := model.selectedProject()
	selectedPipeline, hasPipeline := model.selectedPipeline()
	if !hasProject || !hasPipeline {
		return func() tea.Msg {
			return runsLoadedMessage{requestID: requestID, runs: nil, loadError: nil}
		}
	}

	return func() tea.Msg {
		resultFilter := normalizedRunResultFilter(model.runsResultFilter)
		runs, loadError := model.devopsService.ListRuns(selectedProject.Name, selectedPipeline.ID, defaultRunFetchLimit, resultFilter)
		return runsLoadedMessage{requestID: requestID, runs: runs, loadError: loadError}
	}
}

func (model MainLayoutModel) loadRunDetailsCommand(requestID int) tea.Cmd {
	selectedProject, hasProject := model.selectedProject()
	selectedRun, hasRun := model.selectedRun()
	if !hasProject || !hasRun {
		return func() tea.Msg {
			return runDetailsLoadedMessage{requestID: requestID, runDetails: devops.RunDetails{}, loadError: nil}
		}
	}

	return func() tea.Msg {
		runDetails, loadError := model.devopsService.GetRunDetailsBuildFirst(selectedProject.Name, selectedRun.ID)
		if loadError != nil {
			return runDetailsLoadedMessage{requestID: requestID, runDetails: runDetails, loadError: loadError}
		}

		timeline, timelineLoadError := model.devopsService.GetBuildTimeline(selectedProject.Name, selectedRun.ID)
		logs, logsLoadError := model.devopsService.ListBuildLogs(selectedProject.Name, selectedRun.ID)

		return runDetailsLoadedMessage{
			requestID:         requestID,
			runDetails:        runDetails,
			timeline:          timeline,
			logs:              logs,
			timelineLoadError: timelineLoadError,
			logsLoadError:     logsLoadError,
			loadError:         nil,
		}
	}
}

func (model *MainLayoutModel) nextRequestID(loadingMessage string) int {
	model.activeRequestID++
	model.isLoading = true
	model.loadingMessage = loadingMessage
	return model.activeRequestID
}

func (model *MainLayoutModel) finishLoading() {
	model.isLoading = false
	model.loadingMessage = ""
	model.lastError = ""
}

func (model *MainLayoutModel) markLoadFailure(loadError error) {
	model.isLoading = false
	model.loadingMessage = ""
	model.lastError = loadError.Error()
}

func (model *MainLayoutModel) resetSearch() {
	model.searchInput.SetValue("")
	model.listCursorIndex = 0
}

func (model *MainLayoutModel) applyInitialInputModeIfNeeded() {
	if model.hasInitializedInputMode {
		return
	}

	model.hasInitializedInputMode = true
	if model.initialMode == inputModeInsert {
		model.setModeForStageEntry(model.currentStage)
		return
	}

	model.setNormalMode()
}

func (model *MainLayoutModel) setModeForStageEntry(targetStage browserStage) {
	if targetStage == stageRuns || targetStage == stageRunDetails {
		model.setNormalMode()
		return
	}

	model.currentMode = inputModeInsert
	model.clearInsertEscapeSequence()
	model.searchInput.Focus()
}

func (model *MainLayoutModel) setNormalMode() {
	model.clearInsertEscapeSequence()
	model.currentMode = inputModeNormal
	model.searchInput.Blur()
	if model.currentStage != stageRunDetails {
		model.listCursorIndex = 0
	}
}

func (model *MainLayoutModel) startInsertEscapeSequence() tea.Cmd {
	model.insertEscapeSequence++
	model.insertEscapePending = true
	sequence := model.insertEscapeSequence
	return tea.Tick(insertEscapeSequenceTimeout, func(time.Time) tea.Msg {
		return insertEscapeTimeoutMessage{sequence: sequence}
	})
}

func (model *MainLayoutModel) flushPendingInsertEscape() {
	if !model.insertEscapePending {
		return
	}
	model.clearInsertEscapeSequence()
	model.searchInput.SetValue(model.searchInput.Value() + "j")
}

func (model *MainLayoutModel) clearInsertEscapeSequence() {
	model.insertEscapePending = false
}

func (model *MainLayoutModel) startSetInsertEscapeSequence() tea.Cmd {
	model.setInsertEscapeSequence++
	model.setInsertEscapePending = true
	sequence := model.setInsertEscapeSequence
	return tea.Tick(insertEscapeSequenceTimeout, func(time.Time) tea.Msg {
		return setInsertEscapeTimeoutMessage{sequence: sequence}
	})
}

func (model *MainLayoutModel) flushPendingSetInsertEscape() {
	if !model.setInsertEscapePending {
		return
	}
	model.clearSetInsertEscapeSequence()
	model.setSearchInput.SetValue(model.setSearchInput.Value() + "j")
	model.setWizardCursor = clampSelection(model.setWizardCursor, len(model.filteredSetOptions()))
}

func (model *MainLayoutModel) clearSetInsertEscapeSequence() {
	model.setInsertEscapePending = false
}
