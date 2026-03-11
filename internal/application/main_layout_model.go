package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"lazydevops/internal/application/components/search"
	"lazydevops/internal/devops"
	"lazydevops/internal/settings"
)

func NewMainLayoutModel(organizationURL string) MainLayoutModel {
	applicationSettings := settings.Current()
	configFilePath, _ := settings.ConfigFilePath()
	resolvedOrganizationURL := strings.TrimSpace(organizationURL)
	if strings.TrimSpace(applicationSettings.Organization) != "" {
		resolvedOrganizationURL = strings.TrimSpace(applicationSettings.Organization)
	}

	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(applicationSettings.Palette.Spinner))

	searchInput := search.NewTextInput("Search: ", "Type to filter", 200, 60)
	commandInput := search.NewTextInput(":", "q", 120, 40)
	setValueInput := search.NewTextInput("Value: ", "Enter value", 200, 60)
	setSearchInput := search.NewTextInput("Search: ", "Type to filter", 200, 60)

	currentInputMode := inputModeNormal
	if strings.EqualFold(applicationSettings.MainLayout.DefaultInputMode, string(inputModeInsert)) {
		currentInputMode = inputModeInsert
		searchInput.Focus()
	} else {
		searchInput.Blur()
	}

	model := MainLayoutModel{
		spinnerModel:               spinnerModel,
		searchInput:                searchInput,
		commandInput:               commandInput,
		devopsService:              devops.NewService(resolvedOrganizationURL),
		palette:                    applicationSettings.Palette,
		mainLayoutSettings:         applicationSettings.MainLayout,
		searchSettings:             applicationSettings.Search,
		logRenderingSettings:       applicationSettings.LogRendering,
		runStatusIcons:             applicationSettings.RunStatusIcons,
		runStatusColors:            applicationSettings.RunStatusColors,
		organizationURL:            resolvedOrganizationURL,
		windowWidth:                140,
		windowHeight:               40,
		currentStage:               stageProjects,
		currentMode:                currentInputMode,
		initialMode:                currentInputMode,
		selectedProjectIndex:       0,
		selectedPipelineIndex:      0,
		selectedRunIndex:           0,
		listCursorIndex:            0,
		isLoading:                  true,
		loadingMessage:             "Loading projects...",
		activeRequestID:            1,
		hasInitializedInputMode:    false,
		isHelpVisible:              false,
		isCommandPaletteVisible:    false,
		isRunsStatusPickerVisible:  false,
		commandError:               "",
		modeBeforeCommand:          currentInputMode,
		configFilePath:             configFilePath,
		isSetWizardVisible:         false,
		setWizardStep:              setWizardStepSelect,
		setOptions:                 nil,
		setWizardCursor:            0,
		setSearchInput:             setSearchInput,
		setWizardMode:              inputModeInsert,
		setSelectedOption:          setOption{},
		setValueInput:              setValueInput,
		setWizardError:             "",
		setWizardSuccess:           "",
		setPendingValue:            "",
		runsStatusCursor:           0,
		runDetailsFocusSection:     0,
		runDetailsTreeCursor:       0,
		runDetailsLogsCursor:       0,
		runDetailsExpandedStageIDs: map[string]bool{},
		runDetailsExpandedJobIDs:   map[string]bool{},
		runDetailsLoadedRecordIDs:  map[string]bool{},
		selectedStageKey:           "",
		selectedJobID:              "",
		selectedTaskID:             "",
		selectedLogID:              "",
		selectedLogContent:         "",
		selectedLogError:           "",
		isLogContentLoading:        false,
		activeLogRequestID:         0,
		activeTimelineRequestID:    0,
		logContentCache:            map[string]string{},
		highlightedLogContentCache: map[string][]string{},
		highlightLogFailureCache:   map[string]bool{},
		runStagePreviews:           map[int]runStagePreview{},
		runsResultFilter:           "",
	}
	model.syncInputWidths()
	return model
}

func (model MainLayoutModel) Init() tea.Cmd {
	return tea.Batch(model.spinnerModel.Tick, model.loadProjectsCommand(model.activeRequestID))
}

func (model *MainLayoutModel) syncInputWidths() {
	availableWidth := model.windowWidth - 2
	if availableWidth < 8 {
		availableWidth = 8
	}

	searchWidth := availableWidth - 24
	if searchWidth < 1 {
		searchWidth = 1
	}
	setInputWidth := availableWidth - 12
	if setInputWidth < 1 {
		setInputWidth = 1
	}
	commandWidth := availableWidth - 8
	if commandWidth < 1 {
		commandWidth = 1
	}

	model.searchInput.Width = searchWidth
	model.setSearchInput.Width = setInputWidth
	model.setValueInput.Width = setInputWidth
	model.commandInput.Width = commandWidth
}

func (model MainLayoutModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMessage := message.(type) {
	case tea.WindowSizeMsg:
		model.windowWidth = typedMessage.Width
		model.windowHeight = typedMessage.Height
		model.syncInputWidths()
		return model, model.prefetchRunStagePreviewsCommand()

	case tea.KeyMsg:
		return model.handleKeyMessage(typedMessage)

	case spinner.TickMsg:
		var spinnerCommand tea.Cmd
		model.spinnerModel, spinnerCommand = model.spinnerModel.Update(typedMessage)
		if model.isLoading {
			return model, spinnerCommand
		}
		return model, nil

	case insertEscapeTimeoutMessage:
		if !model.insertEscapePending || typedMessage.sequence != model.insertEscapeSequence || model.currentMode != inputModeInsert {
			return model, nil
		}
		model.flushPendingInsertEscape()
		return model, nil

	case setInsertEscapeTimeoutMessage:
		if !model.setInsertEscapePending || typedMessage.sequence != model.setInsertEscapeSequence || model.setWizardMode != inputModeInsert {
			return model, nil
		}
		model.flushPendingSetInsertEscape()
		return model, nil

	case projectsLoadedMessage:
		if typedMessage.requestID != model.activeRequestID {
			return model, nil
		}
		if typedMessage.loadError != nil {
			setupModel := NewSetupModel()
			return setupModel, setupModel.Init()
		}

		model.projects = typedMessage.projects
		model.pipelines = nil
		model.runs = nil
		model.clearRunExecutionData()
		model.selectedProjectIndex = clampSelection(model.selectedProjectIndex, len(model.projects))

		defaultProjectIndex, hasDefaultProject := model.defaultProjectIndex()
		if hasDefaultProject {
			model.selectedProjectIndex = defaultProjectIndex
			model.selectedPipelineIndex = 0
			model.selectedRunIndex = 0
			model.currentStage = stagePipelines
			model.setModeForStageEntry(stagePipelines)
			requestID := model.nextRequestID("Loading pipelines...")
			return model, tea.Batch(model.spinnerModel.Tick, model.loadPipelinesCommand(requestID))
		}

		model.currentStage = stageProjects
		model.finishLoading()
		model.resetSearch()
		model.applyInitialInputModeIfNeeded()
		return model, nil

	case pipelinesLoadedMessage:
		if typedMessage.requestID != model.activeRequestID {
			return model, nil
		}
		if typedMessage.loadError != nil {
			model.markLoadFailure(typedMessage.loadError)
			model.pipelines = nil
			model.runs = nil
			model.clearRunExecutionData()
			model.currentStage = stagePipelines
			model.resetSearch()
			return model, nil
		}

		model.pipelines = typedMessage.pipelines
		model.runs = nil
		model.clearRunExecutionData()
		model.selectedPipelineIndex = clampSelection(model.selectedPipelineIndex, len(model.pipelines))
		model.currentStage = stagePipelines
		model.finishLoading()
		model.resetSearch()
		return model, nil

	case runsLoadedMessage:
		if typedMessage.requestID != model.activeRequestID {
			return model, nil
		}
		if typedMessage.loadError != nil {
			model.markLoadFailure(typedMessage.loadError)
			model.runs = nil
			model.clearRunExecutionData()
			model.currentStage = stageRuns
			model.resetSearch()
			return model, nil
		}

		model.runs = sortRunsMostRecent(typedMessage.runs)
		model.clearRunExecutionData()
		model.runStagePreviews = map[int]runStagePreview{}
		model.selectedRunIndex = clampSelection(model.selectedRunIndex, len(model.runs))
		model.currentStage = stageRuns
		model.finishLoading()
		model.resetSearch()
		return model, model.prefetchRunStagePreviewsCommand()

	case runDetailsLoadedMessage:
		if typedMessage.requestID != model.activeRequestID {
			return model, nil
		}
		if typedMessage.loadError != nil {
			model.markLoadFailure(typedMessage.loadError)
			model.clearRunExecutionData()
			model.currentStage = stageRunDetails
			return model, nil
		}

		model.selectedRunDetails = typedMessage.runDetails
		model.hasRunDetails = true
		model.selectedRunTimeline = typedMessage.timeline
		model.selectedRunLogs = typedMessage.logs
		model.runExecutionWarnings = collectRunExecutionWarnings(typedMessage.timelineLoadError, typedMessage.logsLoadError)
		model.runDetailsScrollOffset = 0
		model.runDetailsLoadedRecordIDs = map[string]bool{}
		model.initializeRunDetailsExplorer()
		model.currentStage = stageRunDetails
		model.finishLoading()
		return model, nil

	case runDetailsTimelineExpandedMessage:
		if typedMessage.requestID != model.activeTimelineRequestID {
			return model, nil
		}
		model.isLoading = false
		model.loadingMessage = ""
		if model.runDetailsLoadedRecordIDs == nil {
			model.runDetailsLoadedRecordIDs = map[string]bool{}
		}
		model.runDetailsLoadedRecordIDs[strings.TrimSpace(typedMessage.parentRecordID)] = true
		if typedMessage.loadError != nil {
			model.runExecutionWarnings = append(model.runExecutionWarnings, fmt.Sprintf("timeline details unavailable: %s", typedMessage.loadError.Error()))
			return model, nil
		}
		model.selectedRunTimeline = mergeBuildTimelineRecords(model.selectedRunTimeline, typedMessage.timeline)
		return model, nil

	case runLogLoadedMessage:
		if typedMessage.requestID != model.activeLogRequestID {
			return model, nil
		}
		model.isLogContentLoading = false
		model.selectedLogID = typedMessage.logID
		if model.highlightedLogContentCache == nil {
			model.highlightedLogContentCache = map[string][]string{}
		}
		if model.highlightLogFailureCache == nil {
			model.highlightLogFailureCache = map[string]bool{}
		}
		delete(model.highlightedLogContentCache, typedMessage.logID)
		delete(model.highlightLogFailureCache, typedMessage.logID)
		if typedMessage.loadError != nil {
			model.selectedLogContent = ""
			model.selectedLogError = typedMessage.loadError.Error()
			return model, nil
		}
		model.selectedLogError = ""
		model.selectedLogContent = typedMessage.content
		if model.logContentCache == nil {
			model.logContentCache = map[string]string{}
		}
		model.logContentCache[typedMessage.logID] = typedMessage.content
		return model, nil

	case runStagesLoadedMessage:
		if model.runStagePreviews == nil {
			model.runStagePreviews = map[int]runStagePreview{}
		}
		preview := model.runStagePreviews[typedMessage.runID]
		preview.IsLoading = false
		preview.Statuses = append([]string{}, typedMessage.statuses...)
		if typedMessage.loadError != nil {
			preview.LoadError = typedMessage.loadError.Error()
		} else {
			preview.LoadError = ""
		}
		model.runStagePreviews[typedMessage.runID] = preview
		return model, nil
	}

	return model, nil
}

func (model MainLayoutModel) View() string {
	if model.isSetWizardVisible {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Title))
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Help))
		panelWidth := model.windowWidth - 6
		if panelWidth < 1 {
			panelWidth = 1
		}
		lines := []string{titleStyle.Render("lazydevops"), ""}
		if model.setWizardStep == setWizardStepSelect {
			lines = append(lines, model.setSearchInput.View())
			lines = append(lines, "")
		}
		lines = append(lines, model.renderSetWizard(panelWidth))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("set picker: insert=i/s//, normal=esc, move=j/k, select=enter, close=esc (normal mode)"))
		lines = append(lines, "")
		lines = append(lines, model.renderStatusBar(panelWidth))
		return strings.Join(lines, "\n") + "\n"
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Title))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Error)).Bold(true)

	var lines []string
	lines = append(lines, titleStyle.Render("lazydevops"))
	lines = append(lines, "")
	lines = append(lines, model.searchInput.View())
	if model.lastError != "" {
		lines = append(lines, errorStyle.Render(model.lastError))
	}
	lines = append(lines, "")

	footerLineCount := 4 // blank + help + blank + status bar
	if model.isCommandPaletteVisible {
		footerLineCount += 6 // command palette block + spacer
	}
	if model.isRunsStatusPickerVisible {
		footerLineCount += 8 // status picker block + spacer
	}
	panelBorderLineCount := 2 // top + bottom border
	mainPanelSlackLines := 2
	panelHeight := model.windowHeight - len(lines) - footerLineCount - panelBorderLineCount - mainPanelSlackLines
	if panelHeight < 1 {
		panelHeight = 1
	}
	panelWidth := model.windowWidth - 6
	if panelWidth < 1 {
		panelWidth = 1
	}

	lines = append(lines, model.renderSinglePanel(panelWidth, panelHeight))
	lines = append(lines, "")
	if model.isCommandPaletteVisible {
		lines = append(lines, model.renderCommandPalette(panelWidth))
		lines = append(lines, "")
	}
	if model.isRunsStatusPickerVisible {
		lines = append(lines, model.renderRunsStatusPicker(panelWidth))
		lines = append(lines, "")
	}
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Help))
	lines = append(lines, helpStyle.Render(model.helpText()))
	lines = append(lines, "")
	lines = append(lines, model.renderStatusBar(panelWidth))

	return strings.Join(lines, "\n") + "\n"
}
