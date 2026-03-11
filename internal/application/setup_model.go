package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"lazydevops/internal/devops"
	"lazydevops/internal/onboarding"
	"lazydevops/internal/prerequisites"
	"lazydevops/internal/settings"
)

type setupState string

const (
	setupStateChecking             setupState = "checking"
	setupStateComplete             setupState = "complete"
	setupStateEnteringOrganization setupState = "entering_organization"
	setupStateApplyingOrganization setupState = "applying_organization"
	setupStateLaunchingMainLayout  setupState = "launching_main_layout"
)

type checkCompletedMessage struct {
	checkResult prerequisites.Result
}

type organizationConfiguredMessage struct {
	configurationError error
}

type mainLayoutReadyMessage struct {
	mainLayoutModel MainLayoutModel
	initializeError error
}

type SetupModel struct {
	currentState                     setupState
	spinnerModel                     spinner.Model
	organizationInputModel           textinput.Model
	prerequisiteSvc                  prerequisites.Service
	palette                          settings.Palette
	checkDefinitions                 []prerequisites.Definition
	completedResults                 map[prerequisites.CheckIdentifier]prerequisites.Result
	currentCheckIndex                int
	hasAnyFailedChecks               bool
	organizationInputValidationError string
}

func NewSetupModel() SetupModel {
	applicationSettings := settings.Current()

	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(applicationSettings.Palette.Spinner))
	organizationInputModel := textinput.New()
	organizationInputModel.Placeholder = "https://dev.azure.com/your-org"
	organizationInputModel.Prompt = "> "
	organizationInputModel.CharLimit = 200
	organizationInputModel.Width = 70

	prerequisiteService := prerequisites.NewService()
	return SetupModel{
		currentState:           setupStateChecking,
		spinnerModel:           spinnerModel,
		organizationInputModel: organizationInputModel,
		prerequisiteSvc:        prerequisiteService,
		palette:                applicationSettings.Palette,
		checkDefinitions:       prerequisiteService.Definitions(),
		completedResults:       make(map[prerequisites.CheckIdentifier]prerequisites.Result),
		currentCheckIndex:      0,
	}
}

func (model SetupModel) Init() tea.Cmd {
	return tea.Batch(model.spinnerModel.Tick, model.runCurrentCheckCommand())
}

func (model SetupModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMessage := message.(type) {
	case tea.KeyMsg:
		return model.handleKeyMessage(typedMessage)

	case spinner.TickMsg:
		var spinnerCommand tea.Cmd
		model.spinnerModel, spinnerCommand = model.spinnerModel.Update(typedMessage)
		if model.currentState == setupStateChecking || model.currentState == setupStateApplyingOrganization {
			return model, spinnerCommand
		}
		return model, nil

	case checkCompletedMessage:
		model.completedResults[typedMessage.checkResult.Identifier] = typedMessage.checkResult
		if typedMessage.checkResult.Status == prerequisites.StatusFailed {
			model.hasAnyFailedChecks = true
		}

		model.currentCheckIndex++
		if model.currentCheckIndex >= len(model.checkDefinitions) {
			model.currentState = setupStateComplete
			if !model.hasAnyFailedChecks {
				if markError := onboarding.MarkCompleted(); markError != nil {
					model.organizationInputValidationError = fmt.Sprintf("Failed to persist onboarding state: %v", markError)
				}
				model.currentState = setupStateLaunchingMainLayout
				return model, tea.Batch(model.spinnerModel.Tick, model.initializeMainLayoutCommand())
			}
			return model, nil
		}

		return model, model.runCurrentCheckCommand()

	case organizationConfiguredMessage:
		if typedMessage.configurationError != nil {
			model.currentState = setupStateEnteringOrganization
			model.organizationInputValidationError = typedMessage.configurationError.Error()
			model.organizationInputModel.Focus()
			return model, nil
		}

		model.organizationInputValidationError = ""
		return model.resetChecks()

	case mainLayoutReadyMessage:
		if typedMessage.initializeError != nil {
			model.currentState = setupStateComplete
			model.hasAnyFailedChecks = true
			model.organizationInputValidationError = typedMessage.initializeError.Error()
			return model, nil
		}

		return typedMessage.mainLayoutModel, typedMessage.mainLayoutModel.Init()
	}

	return model, nil
}

func (model SetupModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Title))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Help))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Info))
	failedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Error)).Bold(true)
	passedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Success)).Bold(true)
	inputTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Accent))

	var lines []string
	lines = append(lines, titleStyle.Render("lazydevops"))
	lines = append(lines, "")

	if model.currentState == setupStateChecking {
		lines = append(lines, fmt.Sprintf("%s Checking prerequisites...", model.spinnerModel.View()))
		lines = append(lines, "")
	}
	if model.currentState == setupStateApplyingOrganization {
		lines = append(lines, fmt.Sprintf("%s Applying default Azure DevOps organization...", model.spinnerModel.View()))
		lines = append(lines, "")
	}
	if model.currentState == setupStateLaunchingMainLayout {
		lines = append(lines, fmt.Sprintf("%s Loading main layout...", model.spinnerModel.View()))
		lines = append(lines, "")
	}

	if model.currentState == setupStateComplete {
		if model.hasAnyFailedChecks {
			lines = append(lines, failedStyle.Render("Setup incomplete: one or more checks failed."))
		} else {
			lines = append(lines, passedStyle.Render("Setup ready: all checks passed."))
		}
		lines = append(lines, "")
	}
	if model.currentState == setupStateEnteringOrganization {
		lines = append(lines, inputTitleStyle.Render("Set default Azure DevOps organization"))
		lines = append(lines, "Enter your Azure DevOps org URL (for example: https://dev.azure.com/your-org)")
		lines = append(lines, model.organizationInputModel.View())
		lines = append(lines, "")
	}
	if model.organizationInputValidationError != "" {
		lines = append(lines, failedStyle.Render(model.organizationInputValidationError))
		lines = append(lines, "")
	}

	for _, currentDefinition := range model.checkDefinitions {
		lines = append(lines, model.renderCheckLine(currentDefinition))

		checkResult, hasResult := model.completedResults[currentDefinition.Identifier]
		if hasResult && checkResult.Details != "" {
			lines = append(lines, infoStyle.Render("  "+checkResult.Details))
		}

		if hasResult && len(checkResult.Remediation) > 0 {
			for _, remediationLine := range checkResult.Remediation {
				lines = append(lines, infoStyle.Render("  - "+remediationLine))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render(model.helpText()))

	return strings.Join(lines, "\n") + "\n"
}

func (model SetupModel) resetChecks() (tea.Model, tea.Cmd) {
	model.currentState = setupStateChecking
	model.currentCheckIndex = 0
	model.hasAnyFailedChecks = false
	model.completedResults = make(map[prerequisites.CheckIdentifier]prerequisites.Result)
	model.organizationInputValidationError = ""
	return model, tea.Batch(model.spinnerModel.Tick, model.runCurrentCheckCommand())
}

func (model SetupModel) runCurrentCheckCommand() tea.Cmd {
	if model.currentCheckIndex >= len(model.checkDefinitions) {
		return nil
	}

	checkDefinition := model.checkDefinitions[model.currentCheckIndex]
	completedResultsSnapshot := make(map[prerequisites.CheckIdentifier]prerequisites.Result, len(model.completedResults))
	for checkIdentifier, checkResult := range model.completedResults {
		completedResultsSnapshot[checkIdentifier] = checkResult
	}

	return func() tea.Msg {
		checkResult := model.prerequisiteSvc.Run(checkDefinition.Identifier, completedResultsSnapshot)
		return checkCompletedMessage{checkResult: checkResult}
	}
}

func (model SetupModel) renderCheckLine(checkDefinition prerequisites.Definition) string {
	checkResult, hasResult := model.completedResults[checkDefinition.Identifier]

	status := prerequisites.StatusPending
	if hasResult {
		status = checkResult.Status
	} else if model.currentState == setupStateChecking && model.checkDefinitions[model.currentCheckIndex].Identifier == checkDefinition.Identifier {
		status = prerequisites.StatusRunning
	}

	statusText := "PENDING"
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))

	switch status {
	case prerequisites.StatusRunning:
		statusText = "RUNNING"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Spinner)).Bold(true)
	case prerequisites.StatusPassed:
		statusText = "PASS"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Success)).Bold(true)
	case prerequisites.StatusFailed:
		statusText = "FAIL"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Error)).Bold(true)
	}

	return fmt.Sprintf("[%s] %s", statusStyle.Render(statusText), checkDefinition.Title)
}

func (model SetupModel) handleKeyMessage(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMessage.String() == "ctrl+c" {
		return model, tea.Quit
	}

	switch model.currentState {
	case setupStateEnteringOrganization:
		return model.handleOrganizationInputKeyMessage(keyMessage)
	case setupStateChecking, setupStateApplyingOrganization, setupStateLaunchingMainLayout:
		return model, nil
	case setupStateComplete:
		switch keyMessage.String() {
		case "r":
			return model.resetChecks()
		case "o":
			if !model.canEnterOrganizationInput() {
				return model, nil
			}
			model.currentState = setupStateEnteringOrganization
			model.organizationInputValidationError = ""
			model.organizationInputModel.Focus()
			return model, nil
		}
	}

	return model, nil
}

func (model SetupModel) handleOrganizationInputKeyMessage(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMessage.String() {
	case "esc":
		model.currentState = setupStateComplete
		model.organizationInputValidationError = ""
		model.organizationInputModel.Blur()
		return model, nil
	case "enter":
		organizationURL := strings.TrimSpace(model.organizationInputModel.Value())
		validationError := validateOrganizationURL(organizationURL)
		if validationError != "" {
			model.organizationInputValidationError = validationError
			return model, nil
		}

		model.currentState = setupStateApplyingOrganization
		model.organizationInputValidationError = ""
		model.organizationInputModel.Blur()
		return model, tea.Batch(model.spinnerModel.Tick, model.configureOrganizationCommand(organizationURL))
	}

	var updateCommand tea.Cmd
	model.organizationInputModel, updateCommand = model.organizationInputModel.Update(keyMessage)
	return model, updateCommand
}

func (model SetupModel) configureOrganizationCommand(organizationURL string) tea.Cmd {
	return func() tea.Msg {
		configurationError := model.prerequisiteSvc.ConfigureDefaultOrganization(organizationURL)
		return organizationConfiguredMessage{configurationError: configurationError}
	}
}

func (model SetupModel) canEnterOrganizationInput() bool {
	defaultsCheckResult, hasDefaultsCheckResult := model.completedResults[prerequisites.CheckAzureDevOpsDefaultsConfigured]
	return hasDefaultsCheckResult && defaultsCheckResult.Status == prerequisites.StatusFailed
}

func (model SetupModel) helpText() string {
	switch model.currentState {
	case setupStateEnteringOrganization:
		return "enter: apply  esc: cancel"
	case setupStateApplyingOrganization, setupStateLaunchingMainLayout:
		return "checking..."
	case setupStateComplete:
		if model.canEnterOrganizationInput() {
			return "r: recheck  o: set default org"
		}
		return "r: recheck"
	default:
		return ""
	}
}

func (model SetupModel) initializeMainLayoutCommand() tea.Cmd {
	return func() tea.Msg {
		organizationURL, resolveError := devops.ResolveDefaultOrganization()
		if resolveError != nil {
			return mainLayoutReadyMessage{initializeError: resolveError}
		}

		mainLayoutModel := NewMainLayoutModel(organizationURL)
		return mainLayoutReadyMessage{mainLayoutModel: mainLayoutModel}
	}
}

func validateOrganizationURL(organizationURL string) string {
	if organizationURL == "" {
		return "Organization URL is required."
	}
	if !strings.HasPrefix(organizationURL, "https://dev.azure.com/") {
		return "Organization URL must start with https://dev.azure.com/."
	}

	pathSuffix := strings.TrimPrefix(organizationURL, "https://dev.azure.com/")
	pathSuffix = strings.TrimSpace(pathSuffix)
	if pathSuffix == "" || pathSuffix == "/" {
		return "Organization URL must include an organization name."
	}

	return ""
}
