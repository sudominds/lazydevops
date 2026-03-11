package application

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"lazydevops/internal/devops"
	"lazydevops/internal/settings"
)

const defaultRunFetchLimit = 50
const insertEscapeSequenceTimeout = 300 * time.Millisecond

var allowedRunResultFilters = []string{
	"canceled",
	"failed",
	"none",
	"partiallySucceeded",
	"succeeded",
}

type browserStage string
type inputMode string

const (
	stageProjects   browserStage = "projects"
	stagePipelines  browserStage = "pipelines"
	stageRuns       browserStage = "runs"
	stageRunDetails browserStage = "run_details"

	inputModeNormal inputMode = "normal"
	inputModeInsert inputMode = "insert"
)

type projectsLoadedMessage struct {
	requestID int
	projects  []devops.Project
	loadError error
}

type pipelinesLoadedMessage struct {
	requestID int
	pipelines []devops.Pipeline
	loadError error
}

type runsLoadedMessage struct {
	requestID int
	runs      []devops.Run
	loadError error
}

type runDetailsLoadedMessage struct {
	requestID         int
	runDetails        devops.RunDetails
	timeline          devops.BuildTimeline
	logs              []devops.BuildLog
	timelineLoadError error
	logsLoadError     error
	loadError         error
}

type runDetailsTimelineExpandedMessage struct {
	requestID      int
	parentRecordID string
	timeline       devops.BuildTimeline
	loadError      error
}

type runLogLoadedMessage struct {
	requestID  int
	logID      string
	content    string
	loadError  error
	commandCLI string
}

type runStagesLoadedMessage struct {
	runID     int
	statuses  []string
	loadError error
}

type insertEscapeTimeoutMessage struct {
	sequence int
}

type setInsertEscapeTimeoutMessage struct {
	sequence int
}

type setWizardStep string

const (
	setWizardStepSelect  setWizardStep = "select"
	setWizardStepInput   setWizardStep = "input"
	setWizardStepConfirm setWizardStep = "confirm"
)

type setOption struct {
	Key          string
	CurrentValue string
}

type listCard struct {
	Lines []listCardLine
}

type listCardLine struct {
	Text  string
	Muted bool
}

type runStagePreview struct {
	Statuses  []string
	IsLoading bool
	LoadError string
}

type stageSummary struct {
	ID       string
	Name     string
	Status   string
	Duration string
	Order    float64
	HasOrder bool
	Sequence int
	StartAt  string
}

type jobSummary struct {
	ID       string
	ParentID string
	Name     string
	Status   string
	Duration string
	Attempt  string
	LogID    string
	Order    float64
	HasOrder bool
	Sequence int
	StartAt  string
}

type taskSummary struct {
	ID          string
	ParentJobID string
	Name        string
	Status      string
	Duration    string
	Attempt     string
	LogID       string
	Order       float64
	HasOrder    bool
	Sequence    int
	StartAt     string
}

type lineSpan struct {
	Start int
	End   int
}

type runDetailsJobGroup struct {
	Job   jobSummary
	Tasks []taskSummary
}

type runDetailsStageGroup struct {
	Key   string
	Stage stageSummary
	Jobs  []runDetailsJobGroup
}

type runDetailsTreeItemKind string

const (
	runDetailsTreeItemStage runDetailsTreeItemKind = "stage"
	runDetailsTreeItemJob   runDetailsTreeItemKind = "job"
	runDetailsTreeItemTask  runDetailsTreeItemKind = "task"
)

type runDetailsTreeItem struct {
	Kind     runDetailsTreeItemKind
	StageKey string
	Stage    stageSummary
	Job      jobSummary
	Task     taskSummary
}

type MainLayoutModel struct {
	spinnerModel         spinner.Model
	searchInput          textinput.Model
	commandInput         textinput.Model
	devopsService        devops.Service
	palette              settings.Palette
	mainLayoutSettings   settings.MainLayoutSettings
	searchSettings       settings.SearchSettings
	logRenderingSettings settings.LogRenderingSettings
	runStatusIcons       settings.RunStatusIconSettings
	runStatusColors      settings.RunStatusColorSettings

	organizationURL string
	windowWidth     int
	windowHeight    int

	currentStage browserStage
	currentMode  inputMode
	initialMode  inputMode

	projects  []devops.Project
	pipelines []devops.Pipeline
	runs      []devops.Run

	selectedProjectIndex       int
	selectedPipelineIndex      int
	selectedRunIndex           int
	selectedRunDetails         devops.RunDetails
	hasRunDetails              bool
	selectedRunTimeline        devops.BuildTimeline
	selectedRunLogs            []devops.BuildLog
	runExecutionWarnings       []string
	runDetailsScrollOffset     int
	runDetailsFocusSection     int
	runDetailsTreeCursor       int
	runDetailsLogsCursor       int
	runDetailsExpandedStageIDs map[string]bool
	runDetailsExpandedJobIDs   map[string]bool
	runDetailsLoadedRecordIDs  map[string]bool
	selectedStageKey           string
	selectedJobID              string
	selectedTaskID             string
	selectedLogID              string
	selectedLogContent         string
	selectedLogError           string
	isLogContentLoading        bool
	activeLogRequestID         int
	activeTimelineRequestID    int
	logContentCache            map[string]string
	highlightedLogContentCache map[string][]string
	highlightLogFailureCache   map[string]bool
	runStagePreviews           map[int]runStagePreview
	runsResultFilter           string

	listCursorIndex int

	isLoading                 bool
	loadingMessage            string
	lastError                 string
	activeRequestID           int
	hasInitializedInputMode   bool
	isHelpVisible             bool
	isCommandPaletteVisible   bool
	isRunsStatusPickerVisible bool
	commandError              string
	modeBeforeCommand         inputMode
	configFilePath            string

	isSetWizardVisible bool
	setWizardStep      setWizardStep
	setOptions         []setOption
	setWizardCursor    int
	setSearchInput     textinput.Model
	setWizardMode      inputMode
	setSelectedOption  setOption
	setValueInput      textinput.Model
	setWizardError     string
	setWizardSuccess   string
	setPendingValue    string
	runsStatusCursor   int

	insertEscapePending     bool
	insertEscapeSequence    int
	setInsertEscapePending  bool
	setInsertEscapeSequence int

	runDetailsGoToTopPending bool
}
