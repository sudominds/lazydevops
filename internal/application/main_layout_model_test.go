package application

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"lazydevops/internal/devops"
)

func TestParseRunSearchFilterExtractsStatusesAndText(t *testing.T) {
	filter := parseRunSearchFilter("result:failed api result:succeeded")

	expectedResults := []string{"failed", "succeeded"}
	if !reflect.DeepEqual(filter.ResultValues, expectedResults) {
		t.Fatalf("unexpected results: got %v want %v", filter.ResultValues, expectedResults)
	}
	if filter.TextQuery != "api" {
		t.Fatalf("unexpected text query: got %q want %q", filter.TextQuery, "api")
	}
}

func TestParseRunSearchFilterIgnoresEmptyStatusToken(t *testing.T) {
	filter := parseRunSearchFilter("result: result:failed")

	expectedResults := []string{"failed"}
	if !reflect.DeepEqual(filter.ResultValues, expectedResults) {
		t.Fatalf("unexpected results: got %v want %v", filter.ResultValues, expectedResults)
	}
	if filter.TextQuery != "" {
		t.Fatalf("unexpected text query: got %q want empty", filter.TextQuery)
	}
}

func TestRunMatchesResultFilterUsesRunResult(t *testing.T) {
	run := devops.Run{
		Result: "partiallySucceeded",
		State:  "completed",
	}

	if !runMatchesResultFilter(run, []string{"partiallySucceeded"}) {
		t.Fatalf("expected run to match result filter")
	}
}

func TestRunMatchesResultFilterNoResultFiltersMatchesAnyRun(t *testing.T) {
	run := devops.Run{Result: "failed"}
	if !runMatchesResultFilter(run, nil) {
		t.Fatalf("expected run to match when no results are provided")
	}
}

func TestNormalizedRunResultFilterSupportsAzureValues(t *testing.T) {
	if normalizedRunResultFilter("cancelled") != "canceled" {
		t.Fatalf("expected cancelled to normalize to canceled")
	}
	if normalizedRunResultFilter("partiallysucceeded") != "partiallySucceeded" {
		t.Fatalf("expected partiallysucceeded to normalize to partiallySucceeded")
	}
	if normalizedRunResultFilter("unknown") != "" {
		t.Fatalf("expected unknown values to normalize to empty filter")
	}
}

func TestFilteredSetOptionsUsesSharedTokenizedSearch(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.setOptions = []setOption{
		{Key: "main_layout.default_project", CurrentValue: "stock_ledger"},
		{Key: "palette.title", CurrentValue: "#ffffff"},
	}
	model.setSearchInput.SetValue("st ger")

	filtered := model.filteredSetOptions()
	if len(filtered) != 1 {
		t.Fatalf("expected one filtered set option, got %d", len(filtered))
	}
	if filtered[0].Key != "main_layout.default_project" {
		t.Fatalf("unexpected filtered option: got %s", filtered[0].Key)
	}
}

func TestRunDetailsHKeyMovesFocusBackward(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 1

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsFocusSection != 0 {
		t.Fatalf("expected focus section 0 after pressing h, got %d", typedModel.runDetailsFocusSection)
	}
}

func TestRunDetailsRightArrowMovesFocusForward(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 0

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRight})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsFocusSection != 2 {
		t.Fatalf("expected focus section 2 after pressing right, got %d", typedModel.runDetailsFocusSection)
	}
}

func TestRunDetailsLKeyUsesEnterBehavior(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 0

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsFocusSection != 0 {
		t.Fatalf("expected l to keep tree focus when no selectable items, got %d", typedModel.runDetailsFocusSection)
	}
}

func TestRunDetailsEscDoesNotNavigateBackPage(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.currentStage = stageRunDetails
	model.runDetailsFocusSection = 0

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyEsc})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.currentStage != stageRunDetails {
		t.Fatalf("expected esc to keep current run details stage, got %s", typedModel.currentStage)
	}
}

func TestSetWizardJKExitsInsertMode(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.openSetWizard()

	updatedModel, _ := model.handleSetWizardKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel after j key")
	}
	if !typedModel.setInsertEscapePending {
		t.Fatalf("expected pending set-wizard insert escape after j")
	}

	updatedModel, _ = typedModel.handleSetWizardKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	typedModel, ok = updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel after k key")
	}
	if typedModel.setWizardMode != inputModeNormal {
		t.Fatalf("expected set wizard to exit to normal mode after jk, got %s", typedModel.setWizardMode)
	}
	if typedModel.setSearchInput.Value() != "" {
		t.Fatalf("expected jk escape to not append characters to set filter")
	}
}

func TestCurrentRunDetailsLogsAlwaysReturnsAllRunLogs(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.selectedRunLogs = []devops.BuildLog{
		{ID: devops.LooseString("1"), LineCount: devops.LooseString("10")},
		{ID: devops.LooseString("2"), LineCount: devops.LooseString("20")},
	}
	model.selectedJobID = "job-1"
	model.selectedTaskID = "task-1"

	logs := model.currentRunDetailsLogs()
	if len(logs) != 2 {
		t.Fatalf("expected all run logs to be visible, got %d", len(logs))
	}
}

func TestSetSearchInputTextMatchesMainSearchCopy(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	if model.setSearchInput.Prompt != "Search: " {
		t.Fatalf("expected set search prompt to match main prompt, got %q", model.setSearchInput.Prompt)
	}
	if model.setSearchInput.Placeholder != "Type to filter" {
		t.Fatalf("expected set search placeholder to match main placeholder, got %q", model.setSearchInput.Placeholder)
	}
}

func TestRunDetailsHFromContentFocusDoesNotScrollBeforeChangingFocus(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 2
	model.runDetailsScrollOffset = 10

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsFocusSection != 0 {
		t.Fatalf("expected h to immediately move focus to tree section, got %d", typedModel.runDetailsFocusSection)
	}
	if typedModel.runDetailsScrollOffset != 10 {
		t.Fatalf("expected h to not scroll content before changing focus, got offset %d", typedModel.runDetailsScrollOffset)
	}
}

func TestRunDetailsLogContentCtrlDMovesDownHalfPage(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 2
	model.selectedLogContent = strings.Repeat("line\n", 200)
	model.runDetailsScrollOffset = 0
	expectedDelta := model.runDetailsLogViewportLineCount() / 2
	if expectedDelta < 1 {
		expectedDelta = 1
	}

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyCtrlD})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsScrollOffset != expectedDelta {
		t.Fatalf("expected ctrl+d to move log content down by half page (%d), got offset %d", expectedDelta, typedModel.runDetailsScrollOffset)
	}
}

func TestRunDetailsLogContentCtrlUMovesUpHalfPage(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 2
	model.selectedLogContent = strings.Repeat("line\n", 200)
	expectedDelta := model.runDetailsLogViewportLineCount() / 2
	if expectedDelta < 1 {
		expectedDelta = 1
	}
	model.runDetailsScrollOffset = expectedDelta * 2

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyCtrlU})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsScrollOffset != expectedDelta {
		t.Fatalf("expected ctrl+u to move log content up by half page (%d), got offset %d", expectedDelta, typedModel.runDetailsScrollOffset)
	}
}

func TestRunDetailsLogContentGGGoesToTop(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 2
	model.selectedLogContent = strings.Repeat("line\n", 200)
	model.runDetailsScrollOffset = 4

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsScrollOffset != 4 {
		t.Fatalf("expected first g to only arm gg without moving, got offset %d", typedModel.runDetailsScrollOffset)
	}
	if !typedModel.runDetailsGoToTopPending {
		t.Fatalf("expected first g to arm gg state")
	}

	updatedModel, _ = typedModel.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	typedModel, ok = updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsScrollOffset != 0 {
		t.Fatalf("expected gg to move to top, got offset %d", typedModel.runDetailsScrollOffset)
	}
	if typedModel.runDetailsGoToTopPending {
		t.Fatalf("expected gg state to clear after navigation")
	}
}

func TestRunDetailsLogContentUpperGGoesToBottom(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.runDetailsFocusSection = 2
	model.selectedLogContent = strings.Repeat("line\n", 200)
	model.runDetailsScrollOffset = 0

	updatedModel, _ := model.handleRunDetailsKeyMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	typedModel, ok := updatedModel.(MainLayoutModel)
	if !ok {
		t.Fatalf("expected MainLayoutModel from handleRunDetailsKeyMessage")
	}
	if typedModel.runDetailsScrollOffset <= 0 {
		t.Fatalf("expected G to move to bottom, got offset %d", typedModel.runDetailsScrollOffset)
	}
}
