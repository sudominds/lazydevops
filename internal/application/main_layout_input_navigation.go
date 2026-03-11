package application

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (model MainLayoutModel) handleKeyMessage(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMessage.String() {
	case "ctrl+c":
		return model, tea.Quit
	}

	if model.isSetWizardVisible {
		return model.handleSetWizardKeyMessage(keyMessage)
	}

	if model.isCommandPaletteVisible {
		switch keyMessage.String() {
		case "esc":
			model.closeCommandPalette(true)
			return model, nil
		case "enter":
			return model.executeCommandPalette()
		}

		var commandInputUpdate tea.Cmd
		model.commandInput, commandInputUpdate = model.commandInput.Update(keyMessage)
		return model, commandInputUpdate
	}
	if model.isRunsStatusPickerVisible {
		return model.handleRunsStatusPickerKeyMessage(keyMessage)
	}

	if model.isHelpVisible {
		switch keyMessage.String() {
		case "?", "esc":
			model.isHelpVisible = false
		}
		return model, nil
	}

	if keyMessage.String() == "?" {
		if model.insertEscapePending {
			model.flushPendingInsertEscape()
		}
		model.isHelpVisible = true
		return model, nil
	}
	if keyMessage.String() == ":" {
		if model.insertEscapePending {
			model.flushPendingInsertEscape()
		}
		model.openCommandPalette()
		return model, nil
	}
	if keyMessage.String() == "alt+f" {
		model.openRunsStatusPicker()
		return model, nil
	}

	if model.isLoading {
		return model, nil
	}

	if model.currentMode == inputModeInsert {
		if keyMessage.String() == "esc" || keyMessage.String() == "enter" {
			model.clearInsertEscapeSequence()
			model.setNormalMode()
			return model, nil
		}

		if model.currentStage == stageRunDetails {
			model.clearInsertEscapeSequence()
			model.setNormalMode()
			return model, nil
		}

		if model.insertEscapePending {
			if keyMessage.String() == "k" {
				model.clearInsertEscapeSequence()
				model.setNormalMode()
				return model, nil
			}
			model.flushPendingInsertEscape()
		}

		if keyMessage.String() == "j" {
			return model, model.startInsertEscapeSequence()
		}

		var updateCommand tea.Cmd
		model.searchInput, updateCommand = model.searchInput.Update(keyMessage)
		return model, tea.Batch(updateCommand, model.prefetchRunStagePreviewsCommand())
	}

	if model.currentStage == stageRunDetails {
		return model.handleRunDetailsKeyMessage(keyMessage)
	}

	switch keyMessage.String() {
	case "up", "k":
		model.moveCursor(-1)
		return model, model.prefetchRunStagePreviewsCommand()
	case "down", "j":
		model.moveCursor(1)
		return model, model.prefetchRunStagePreviewsCommand()
	case "enter", "l":
		return model.selectCurrentItem()
	case "esc", "b":
		model.goBackOneStage()
		return model, nil
	case "r":
		return model.refreshCurrentStage()
	case "/", "i", "s":
		if model.currentStage != stageRunDetails {
			model.currentMode = inputModeInsert
			model.clearInsertEscapeSequence()
			model.searchInput.Focus()
		}
		return model, nil
	}

	return model, nil
}

func (model *MainLayoutModel) moveCursor(delta int) {
	visibleItems := model.filteredItemIndexes()
	if len(visibleItems) == 0 {
		model.listCursorIndex = 0
		return
	}
	model.listCursorIndex = clampSelection(model.listCursorIndex+delta, len(visibleItems))
}

func (model MainLayoutModel) handleRunDetailsKeyMessage(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	if model.runDetailsFocusSection == 2 {
		if handled, updatedModel := model.handleRunDetailsLogContentVimMotion(keyMessage); handled {
			return updatedModel, nil
		}
	}
	model.runDetailsGoToTopPending = false

	switch keyMessage.String() {
	case "tab", "right":
		model.cycleRunDetailsFocus(1)
		return model, nil
	case "left":
		model.cycleRunDetailsFocus(-1)
		return model, nil
	case "h":
		return model.handleRunDetailsBack()
	case "0", "1", "2":
		requestedSection := int(keyMessage.String()[0] - '0')
		if requestedSection == 1 {
			requestedSection = 2
		}
		model.runDetailsFocusSection = requestedSection
		return model, nil
	case "up", "k":
		return model.handleRunDetailsMove(-1)
	case "down", "j":
		return model.handleRunDetailsMove(1)
	case "enter", "l":
		return model.handleRunDetailsEnter()
	case "b":
		model.goBackOneStage()
		return model, nil
	case "esc":
		return model, nil
	case "r":
		return model.refreshCurrentStage()
	}

	return model, nil
}

func (model MainLayoutModel) handleRunDetailsLogContentVimMotion(keyMessage tea.KeyMsg) (bool, MainLayoutModel) {
	switch keyMessage.Type {
	case tea.KeyCtrlD:
		model.runDetailsGoToTopPending = false
		model.scrollRunDetailsHalfPage(1)
		return true, model
	case tea.KeyCtrlU:
		model.runDetailsGoToTopPending = false
		model.scrollRunDetailsHalfPage(-1)
		return true, model
	}

	if keyMessage.Type != tea.KeyRunes || len(keyMessage.Runes) != 1 {
		return false, model
	}

	switch keyMessage.Runes[0] {
	case 'g':
		if model.runDetailsGoToTopPending {
			model.scrollRunDetailsToTop()
			model.runDetailsGoToTopPending = false
			return true, model
		}
		model.runDetailsGoToTopPending = true
		return true, model
	case 'G':
		model.runDetailsGoToTopPending = false
		model.scrollRunDetailsToBottom()
		return true, model
	default:
		return false, model
	}
}

func (model MainLayoutModel) selectCurrentItem() (tea.Model, tea.Cmd) {
	visibleItems := model.filteredItemIndexes()
	if len(visibleItems) == 0 {
		return model, nil
	}

	selectedIndex := visibleItems[model.listCursorIndex]

	switch model.currentStage {
	case stageProjects:
		model.selectedProjectIndex = selectedIndex
		model.selectedPipelineIndex = 0
		model.selectedRunIndex = 0
		model.pipelines = nil
		model.runs = nil
		model.clearRunExecutionData()
		model.lastError = ""
		model.currentStage = stagePipelines
		model.setModeForStageEntry(stagePipelines)
		requestID := model.nextRequestID("Loading pipelines...")
		return model, tea.Batch(model.spinnerModel.Tick, model.loadPipelinesCommand(requestID))

	case stagePipelines:
		model.selectedPipelineIndex = selectedIndex
		model.selectedRunIndex = 0
		model.runs = nil
		model.clearRunExecutionData()
		model.lastError = ""
		model.setModeForStageEntry(stageRuns)
		requestID := model.nextRequestID(fmt.Sprintf("Loading runs (result: %s)...", model.activeRunResultFilterLabel()))
		return model, tea.Batch(model.spinnerModel.Tick, model.loadRunsCommand(requestID))

	case stageRuns:
		model.selectedRunIndex = selectedIndex
		model.clearRunExecutionData()
		model.lastError = ""
		requestID := model.nextRequestID("Loading build details...")
		return model, tea.Batch(model.spinnerModel.Tick, model.loadRunDetailsCommand(requestID))
	}

	return model, nil
}

func (model *MainLayoutModel) goBackOneStage() {
	model.setNormalMode()

	switch model.currentStage {
	case stagePipelines:
		model.currentStage = stageProjects
		model.resetSearch()
	case stageRuns:
		model.currentStage = stagePipelines
		model.resetSearch()
	case stageRunDetails:
		model.currentStage = stageRuns
		model.runDetailsScrollOffset = 0
		model.resetSearch()
	}
}

func (model MainLayoutModel) refreshCurrentStage() (tea.Model, tea.Cmd) {
	model.lastError = ""

	switch model.currentStage {
	case stageProjects:
		requestID := model.nextRequestID("Loading projects...")
		return model, tea.Batch(model.spinnerModel.Tick, model.loadProjectsCommand(requestID))
	case stagePipelines:
		requestID := model.nextRequestID("Loading pipelines...")
		return model, tea.Batch(model.spinnerModel.Tick, model.loadPipelinesCommand(requestID))
	case stageRuns:
		requestID := model.nextRequestID(fmt.Sprintf("Loading runs (result: %s)...", model.activeRunResultFilterLabel()))
		return model, tea.Batch(model.spinnerModel.Tick, model.loadRunsCommand(requestID))
	case stageRunDetails:
		requestID := model.nextRequestID("Loading build details...")
		return model, tea.Batch(model.spinnerModel.Tick, model.loadRunDetailsCommand(requestID))
	default:
		requestID := model.nextRequestID("Loading projects...")
		return model, tea.Batch(model.spinnerModel.Tick, model.loadProjectsCommand(requestID))
	}
}

func (model *MainLayoutModel) openRunsStatusPicker() {
	if model.currentStage != stageRuns {
		return
	}
	model.isRunsStatusPickerVisible = true
	model.currentMode = inputModeNormal
	model.searchInput.Blur()
	model.clearInsertEscapeSequence()
	model.runsStatusCursor = 0
	for index, value := range allowedRunResultFilters {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(model.runsResultFilter)) {
			model.runsStatusCursor = index
			break
		}
	}
}

func (model *MainLayoutModel) closeRunsStatusPicker() {
	model.isRunsStatusPickerVisible = false
	model.runsStatusCursor = 0
}

func (model MainLayoutModel) handleRunsStatusPickerKeyMessage(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMessage.String() {
	case "esc":
		model.closeRunsStatusPicker()
		return model, nil
	case "up", "k":
		model.runsStatusCursor = clampSelection(model.runsStatusCursor-1, len(allowedRunResultFilters))
		return model, nil
	case "down", "j":
		model.runsStatusCursor = clampSelection(model.runsStatusCursor+1, len(allowedRunResultFilters))
		return model, nil
	case "enter":
		selectedValue := allowedRunResultFilters[clampSelection(model.runsStatusCursor, len(allowedRunResultFilters))]
		model.closeRunsStatusPicker()
		if strings.EqualFold(strings.TrimSpace(selectedValue), strings.TrimSpace(model.runsResultFilter)) {
			return model, nil
		}
		model.runsResultFilter = selectedValue
		model.runStagePreviews = map[int]runStagePreview{}
		requestID := model.nextRequestID(fmt.Sprintf("Loading runs (result: %s)...", model.activeRunResultFilterLabel()))
		return model, tea.Batch(model.spinnerModel.Tick, model.loadRunsCommand(requestID))
	}
	return model, nil
}

func (model *MainLayoutModel) prefetchRunStagePreviewsCommand() tea.Cmd {
	if model.currentStage != stageRuns {
		return nil
	}
	selectedProject, hasProject := model.selectedProject()
	if !hasProject {
		return nil
	}
	filteredIndexes := model.filteredItemIndexes()
	if len(filteredIndexes) == 0 {
		return nil
	}

	if model.runStagePreviews == nil {
		model.runStagePreviews = map[int]runStagePreview{}
	}

	cursor := clampSelection(model.listCursorIndex, len(filteredIndexes))
	start := cursor - 2
	if start < 0 {
		start = 0
	}
	end := cursor + 3
	if end >= len(filteredIndexes) {
		end = len(filteredIndexes) - 1
	}

	loadCommands := make([]tea.Cmd, 0, end-start+1)
	for visibleIndex := start; visibleIndex <= end; visibleIndex++ {
		runIndex := filteredIndexes[visibleIndex]
		if runIndex < 0 || runIndex >= len(model.runs) {
			continue
		}
		runID := model.runs[runIndex].ID
		if runID <= 0 {
			continue
		}
		previewState, hasPreview := model.runStagePreviews[runID]
		if hasPreview && (previewState.IsLoading || len(previewState.Statuses) > 0 || strings.TrimSpace(previewState.LoadError) != "") {
			continue
		}
		previewState.IsLoading = true
		previewState.Statuses = nil
		previewState.LoadError = ""
		model.runStagePreviews[runID] = previewState
		loadCommands = append(loadCommands, model.loadRunStagesCommand(selectedProject.Name, runID))
	}

	if len(loadCommands) == 0 {
		return nil
	}
	return tea.Batch(loadCommands...)
}

func (model MainLayoutModel) loadRunStagesCommand(projectName string, runID int) tea.Cmd {
	return func() tea.Msg {
		timeline, loadError := model.devopsService.GetBuildTimeline(projectName, runID)
		if loadError != nil {
			return runStagesLoadedMessage{runID: runID, statuses: nil, loadError: loadError}
		}
		stages, _, _ := summarizeExecution(timeline)
		stageStatuses := make([]string, 0, len(stages))
		for _, stage := range stages {
			stageStatuses = append(stageStatuses, stage.Status)
		}
		return runStagesLoadedMessage{runID: runID, statuses: stageStatuses}
	}
}
