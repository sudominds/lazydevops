package application

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (model MainLayoutModel) renderCommandPalette(panelWidth int) string {
	if model.isSetWizardVisible {
		return model.renderSetWizard(panelWidth)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Accent))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Error))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))

	lines := []string{
		titleStyle.Render("Command"),
		model.commandInput.View(),
		mutedStyle.Render("Examples: :q | :quit | :r | :help | :set"),
	}
	if model.commandError != "" {
		lines = append(lines, errorStyle.Render(model.commandError))
	}

	boxWidth := panelWidth / 2
	if boxWidth < 50 {
		boxWidth = 50
	}
	if boxWidth > panelWidth {
		boxWidth = panelWidth
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(model.palette.Accent)).
		Width(boxWidth).
		Padding(0, 1)

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func (model MainLayoutModel) renderRunsStatusPicker(panelWidth int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Accent))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.SelectedForeground)).Background(lipgloss.Color(model.palette.SelectedBackground))

	lines := []string{
		titleStyle.Render("Runs Result Filter"),
		mutedStyle.Render("Select result and press enter"),
		"",
	}
	selectedIndex := clampSelection(model.runsStatusCursor, len(allowedRunResultFilters))
	for index, statusValue := range allowedRunResultFilters {
		prefix := "  "
		if index == selectedIndex {
			prefix = "> "
		}
		label := fmt.Sprintf("%s%s", prefix, statusValue)
		if index == selectedIndex {
			label = selectedStyle.Render(label)
		}
		lines = append(lines, label)
	}
	lines = append(lines, "", mutedStyle.Render("move: j/k or up/down | apply: enter | close: esc"))

	boxWidth := panelWidth / 2
	if boxWidth < 40 {
		boxWidth = 40
	}
	if boxWidth > panelWidth {
		boxWidth = panelWidth
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(model.palette.Accent)).
		Width(boxWidth).
		Padding(0, 1)

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func (model *MainLayoutModel) openCommandPalette() {
	model.modeBeforeCommand = model.currentMode
	model.clearInsertEscapeSequence()
	model.currentMode = inputModeNormal
	model.searchInput.Blur()
	model.isSetWizardVisible = false
	model.commandInput.SetValue("")
	model.commandInput.Focus()
	model.commandError = ""
	model.isCommandPaletteVisible = true
}

func (model *MainLayoutModel) closeCommandPalette(restorePreviousMode bool) {
	model.isCommandPaletteVisible = false
	model.isSetWizardVisible = false
	model.commandInput.Blur()
	model.commandInput.SetValue("")
	model.commandError = ""
	model.setWizardError = ""
	model.setWizardSuccess = ""
	model.setPendingValue = ""
	model.setValueInput.Blur()

	if !restorePreviousMode {
		return
	}
	if model.currentStage != stageRunDetails && model.modeBeforeCommand == inputModeInsert {
		model.currentMode = inputModeInsert
		model.clearInsertEscapeSequence()
		model.searchInput.Focus()
		return
	}

	model.currentMode = inputModeNormal
	model.searchInput.Blur()
}

func (model MainLayoutModel) executeCommandPalette() (tea.Model, tea.Cmd) {
	commandValue := strings.TrimSpace(model.commandInput.Value())
	commandValue = strings.TrimPrefix(commandValue, ":")
	commandValue = strings.ToLower(strings.TrimSpace(commandValue))

	switch commandValue {
	case "q", "quit", "exit":
		return model, tea.Quit
	case "help", "?":
		model.closeCommandPalette(false)
		model.isHelpVisible = true
		return model, nil
	case "r", "refresh":
		model.closeCommandPalette(false)
		return model.refreshCurrentStage()
	case "i", "insert":
		model.closeCommandPalette(false)
		if model.currentStage != stageRunDetails {
			model.currentMode = inputModeInsert
			model.clearInsertEscapeSequence()
			model.searchInput.Focus()
		}
		return model, nil
	case "n", "normal":
		model.closeCommandPalette(false)
		model.currentMode = inputModeNormal
		model.searchInput.Blur()
		return model, nil
	case "b", "back":
		model.closeCommandPalette(false)
		model.goBackOneStage()
		return model, nil
	case "set":
		model.openSetWizard()
		return model, nil
	case "":
		model.commandError = "command is empty"
		return model, nil
	default:
		model.commandError = fmt.Sprintf("unknown command: %s", commandValue)
		return model, nil
	}
}

func (model *MainLayoutModel) openSetWizard() {
	model.isCommandPaletteVisible = false
	model.commandInput.Blur()
	model.commandInput.SetValue("")
	model.commandError = ""
	model.isSetWizardVisible = true
	model.setWizardStep = setWizardStepSelect
	model.setOptions = model.buildSetOptions()
	model.setWizardCursor = 0
	model.setSearchInput.SetValue("")
	model.setWizardMode = inputModeInsert
	model.clearSetInsertEscapeSequence()
	model.setSearchInput.Focus()
	model.setSelectedOption = setOption{}
	model.setValueInput.SetValue("")
	model.setValueInput.Blur()
	model.setWizardError = ""
	model.setWizardSuccess = ""
	model.setPendingValue = ""
	model.commandError = ""
}

func (model MainLayoutModel) renderSetWizard(panelWidth int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Accent))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Error))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Success))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))

	lines := []string{titleStyle.Render("Set Configuration")}
	lines = append(lines, mutedStyle.Render("Config file: "+truncateWithEllipsis(configuredValue(model.configFilePath), panelWidth-10)))
	lines = append(lines, "")
	switch model.setWizardStep {
	case setWizardStepSelect:
		modeLabel := "NORMAL"
		if model.setWizardMode == inputModeInsert {
			modeLabel = "INSERT"
		}
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("Choose a setting [%s]", modeLabel)))
		lines = append(lines, mutedStyle.Render("insert: type filter | normal: j/k move, enter select"))
		filteredOptions := model.filteredSetOptions()
		if len(filteredOptions) == 0 {
			lines = append(lines, mutedStyle.Render("No matching settings."))
			break
		}
		selectedIndex := clampSelection(model.setWizardCursor, len(filteredOptions))
		maxOptionRows := model.windowHeight - 18
		if maxOptionRows < 4 {
			maxOptionRows = 4
		}
		if maxOptionRows > len(filteredOptions) {
			maxOptionRows = len(filteredOptions)
		}
		startIndex := selectedIndex - (maxOptionRows / 2)
		if startIndex < 0 {
			startIndex = 0
		}
		if startIndex > len(filteredOptions)-maxOptionRows {
			startIndex = len(filteredOptions) - maxOptionRows
		}
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := startIndex + maxOptionRows
		if endIndex > len(filteredOptions) {
			endIndex = len(filteredOptions)
		}
		for index := startIndex; index < endIndex; index++ {
			option := filteredOptions[index]
			line := fmt.Sprintf("%s = %s", option.Key, option.CurrentValue)
			showSelection := model.setWizardMode == inputModeNormal
			lines = append(lines, model.renderSetSelectableLine(index == selectedIndex && showSelection, line, model.setSearchInput.Value(), index))
		}
		if len(filteredOptions) > maxOptionRows {
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("Showing %d-%d of %d", startIndex+1, endIndex, len(filteredOptions))))
		}
	case setWizardStepInput:
		lines = append(lines, mutedStyle.Render("Enter value and press enter:"))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("key: %s", model.setSelectedOption.Key)))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("current: %s", model.setSelectedOption.CurrentValue)))
		allowedValues := setKeyAllowedValues(model.setSelectedOption.Key)
		if len(allowedValues) > 0 {
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("allowed: %s", strings.Join(allowedValues, " | "))))
		}
		lines = append(lines, model.setValueInput.View())
	case setWizardStepConfirm:
		lines = append(lines, mutedStyle.Render("Confirm change (enter=yes, esc=no):"))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("key: %s", model.setSelectedOption.Key)))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("current: %s", model.setSelectedOption.CurrentValue)))
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("new: %s", model.setPendingValue)))
	}

	if model.setWizardError != "" {
		lines = append(lines, "", errorStyle.Render(model.setWizardError))
	}
	if model.setWizardSuccess != "" {
		lines = append(lines, "", successStyle.Render(model.setWizardSuccess))
	}
	lines = append(lines, "", mutedStyle.Render("esc: back"))

	boxWidth := panelWidth - 4
	if boxWidth < 20 {
		boxWidth = 20
	}
	if boxWidth > panelWidth {
		boxWidth = panelWidth
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(model.palette.Accent)).
		Width(boxWidth).
		Padding(0, 1)

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func (model MainLayoutModel) handleSetWizardKeyMessage(keyMessage tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch model.setWizardStep {
	case setWizardStepSelect:
		filteredOptions := model.filteredSetOptions()
		if model.setWizardMode == inputModeInsert {
			switch keyMessage.String() {
			case "esc":
				model.clearSetInsertEscapeSequence()
				model.setWizardMode = inputModeNormal
				model.setSearchInput.Blur()
				return model, nil
			case "enter":
				model.clearSetInsertEscapeSequence()
				if len(filteredOptions) == 0 {
					return model, nil
				}
				selectedOption := filteredOptions[clampSelection(model.setWizardCursor, len(filteredOptions))]
				model.setSelectedOption = selectedOption
				model.setValueInput.SetValue("")
				model.setValueInput.Focus()
				model.setSearchInput.Blur()
				model.setWizardError = ""
				model.setWizardSuccess = ""
				model.setWizardStep = setWizardStepInput
				return model, nil
			case "up":
				model.clearSetInsertEscapeSequence()
				model.setWizardCursor = clampSelection(model.setWizardCursor-1, len(filteredOptions))
				return model, nil
			case "down":
				model.clearSetInsertEscapeSequence()
				model.setWizardCursor = clampSelection(model.setWizardCursor+1, len(filteredOptions))
				return model, nil
			}
			if model.setInsertEscapePending {
				if keyMessage.String() == "k" {
					model.clearSetInsertEscapeSequence()
					model.setWizardMode = inputModeNormal
					model.setSearchInput.Blur()
					return model, nil
				}
				model.flushPendingSetInsertEscape()
			}
			if keyMessage.String() == "j" {
				return model, model.startSetInsertEscapeSequence()
			}
			var updateCommand tea.Cmd
			model.setSearchInput, updateCommand = model.setSearchInput.Update(keyMessage)
			model.setWizardCursor = clampSelection(model.setWizardCursor, len(model.filteredSetOptions()))
			return model, updateCommand
		}

		switch keyMessage.String() {
		case "esc":
			model.isSetWizardVisible = false
			if model.currentStage != stageRunDetails && model.modeBeforeCommand == inputModeInsert {
				model.currentMode = inputModeInsert
				model.clearInsertEscapeSequence()
				model.searchInput.Focus()
			} else {
				model.currentMode = inputModeNormal
				model.searchInput.Blur()
			}
			return model, nil
		case "i", "s", "/":
			model.setWizardMode = inputModeInsert
			model.clearSetInsertEscapeSequence()
			model.setSearchInput.Focus()
			return model, nil
		case "up", "k":
			model.setWizardCursor = clampSelection(model.setWizardCursor-1, len(filteredOptions))
			return model, nil
		case "down", "j":
			model.setWizardCursor = clampSelection(model.setWizardCursor+1, len(filteredOptions))
			return model, nil
		case "enter":
			if len(filteredOptions) == 0 {
				return model, nil
			}
			selectedOption := filteredOptions[clampSelection(model.setWizardCursor, len(filteredOptions))]
			model.setSelectedOption = selectedOption
			model.setValueInput.SetValue("")
			model.setValueInput.Focus()
			model.setSearchInput.Blur()
			model.setWizardError = ""
			model.setWizardSuccess = ""
			model.setWizardStep = setWizardStepInput
			return model, nil
		}
		return model, nil

	case setWizardStepInput:
		switch keyMessage.String() {
		case "esc":
			model.setWizardStep = setWizardStepSelect
			model.setValueInput.Blur()
			model.setWizardMode = inputModeInsert
			model.clearSetInsertEscapeSequence()
			model.setSearchInput.Focus()
			model.setWizardError = ""
			return model, nil
		case "enter":
			inputValue := strings.TrimSpace(model.setValueInput.Value())
			validationError := validateSetValue(model.setSelectedOption.Key, inputValue)
			if validationError != nil {
				model.setWizardError = validationError.Error()
				return model, nil
			}
			model.setPendingValue = inputValue
			model.setWizardError = ""
			model.setWizardStep = setWizardStepConfirm
			model.setValueInput.Blur()
			return model, nil
		}
		var updateCommand tea.Cmd
		model.setValueInput, updateCommand = model.setValueInput.Update(keyMessage)
		return model, updateCommand

	case setWizardStepConfirm:
		switch keyMessage.String() {
		case "esc":
			model.setWizardStep = setWizardStepInput
			model.setValueInput.Focus()
			return model, nil
		case "enter":
			updatedModel, applyError := model.applySetValue()
			if applyError != nil {
				updatedModel.setWizardError = applyError.Error()
				updatedModel.setWizardStep = setWizardStepInput
				updatedModel.setValueInput.Focus()
				return updatedModel, nil
			}
			updatedModel.setWizardSuccess = "updated and saved"
			updatedModel.setWizardError = ""
			updatedModel.setWizardStep = setWizardStepSelect
			updatedModel.setOptions = updatedModel.buildSetOptions()
			updatedModel.setWizardCursor = clampSelection(updatedModel.setWizardCursor, len(updatedModel.filteredSetOptions()))
			updatedModel.setWizardMode = inputModeInsert
			updatedModel.clearSetInsertEscapeSequence()
			updatedModel.setSearchInput.Focus()
			return updatedModel, nil
		}
	}

	return model, nil
}
