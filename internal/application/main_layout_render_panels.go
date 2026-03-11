package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (model MainLayoutModel) renderSinglePanel(panelWidth int, panelHeight int) string {
	if model.isHelpVisible {
		return model.renderHelpPanel(panelWidth, panelHeight)
	}

	title := model.stageTitle()

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(model.palette.Accent)).
		Width(panelWidth).
		Height(panelHeight).
		Padding(0, 1)

	loaderSlot := lipgloss.NewStyle().Width(1)
	loaderGlyph := ""
	if model.isLoading {
		loaderGlyph = model.spinnerModel.View()
	}
	headerLine := loaderSlot.Render(loaderGlyph) + " " + title
	if model.mainLayoutSettings.ShowPathInTitle {
		breadcrumbText := strings.TrimPrefix(model.breadcrumb(), "Path: ")
		headerLine = headerLine + " > " + breadcrumbText
	}
	contentWidth := panelWidth - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	headerLine = fitLineToWidth(headerLine, contentWidth)
	visibleLineCount := panelHeight - 2
	if visibleLineCount < 1 {
		visibleLineCount = 1
	}

	showCommandPreview := visibleLineCount >= 5
	contentLineBudget := visibleLineCount - 2 // header + gap
	if showCommandPreview {
		contentLineBudget = visibleLineCount - 5 // header, gap, bottom gap, separator, command line
	}
	if contentLineBudget < 0 {
		contentLineBudget = 0
	}
	contentLines := model.stageContentLines(contentWidth, contentLineBudget)

	renderedLines := make([]string, 0, visibleLineCount)
	renderedLines = append(renderedLines, headerLine)
	renderedLines = append(renderedLines, "")
	for index := 0; index < contentLineBudget && index < len(contentLines); index++ {
		renderedLines = append(renderedLines, contentLines[index])
	}
	for len(renderedLines) < 2+contentLineBudget {
		renderedLines = append(renderedLines, "")
	}
	if showCommandPreview {
		renderedLines = append(renderedLines, "")
		separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))
		separatorWidth := panelWidth - 4
		if separatorWidth < 1 {
			separatorWidth = 1
		}
		renderedLines = append(renderedLines, separatorStyle.Render(strings.Repeat("─", separatorWidth)))
		commandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))
		renderedLines = append(renderedLines, commandStyle.Render(fitLineToWidth("az: "+model.currentAZCommandPreview(), contentWidth)))
	}
	for len(renderedLines) < visibleLineCount {
		renderedLines = append(renderedLines, "")
	}
	if len(renderedLines) > visibleLineCount {
		renderedLines = renderedLines[:visibleLineCount]
	}

	return panelStyle.Render(strings.Join(renderedLines, "\n"))
}

func (model MainLayoutModel) renderHelpPanel(panelWidth int, panelHeight int) string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(model.palette.Accent)).
		Width(panelWidth).
		Height(panelHeight).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Accent))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(model.palette.Info))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Muted))

	mainLayoutSettings := model.mainLayoutSettings
	configLines := []string{
		formatAlignedLabelValue("organization", configuredValue(model.organizationURL)),
		formatAlignedLabelValue("main_layout.default_input_mode", mainLayoutSettings.DefaultInputMode),
		formatAlignedLabelValue("main_layout.show_path_in_title", fmt.Sprintf("%t", mainLayoutSettings.ShowPathInTitle)),
		formatAlignedLabelValue("main_layout.show_path_in_status_bar", fmt.Sprintf("%t", mainLayoutSettings.ShowPathInStatusBar)),
		formatAlignedLabelValue("main_layout.status_bar_path_side", mainLayoutSettings.StatusBarPathSide),
		formatAlignedLabelValue("main_layout.list_highlight_mode", configuredValue(mainLayoutSettings.ListHighlightMode)),
		formatAlignedLabelValue("main_layout.default_project", configuredValue(mainLayoutSettings.DefaultProject)),
		formatAlignedLabelValue("search.match_highlight_mode", configuredValue(model.searchSettings.MatchHighlightMode)),
		formatAlignedLabelValue("search.rainbow_colors", configuredValue(strings.Join(model.searchSettings.RainbowColors, ","))),
		formatAlignedLabelValue("log_rendering.mode", configuredValue(model.logRenderingSettings.Mode)),
		formatAlignedLabelValue("palette.title", model.palette.Title),
		formatAlignedLabelValue("palette.info", model.palette.Info),
		formatAlignedLabelValue("palette.error", model.palette.Error),
		formatAlignedLabelValue("palette.help", model.palette.Help),
		formatAlignedLabelValue("palette.accent", model.palette.Accent),
		formatAlignedLabelValue("palette.spinner", model.palette.Spinner),
		formatAlignedLabelValue("palette.success", model.palette.Success),
		formatAlignedLabelValue("palette.muted", model.palette.Muted),
		formatAlignedLabelValue("palette.selected_foreground", model.palette.SelectedForeground),
		formatAlignedLabelValue("palette.selected_background", model.palette.SelectedBackground),
		formatAlignedLabelValue("palette.insert_selected_foreground", model.palette.InsertSelectedForeground),
		formatAlignedLabelValue("palette.insert_selected_background", model.palette.InsertSelectedBackground),
		formatAlignedLabelValue("palette.status_bar_foreground", model.palette.StatusBarForeground),
		formatAlignedLabelValue("palette.status_bar_background", model.palette.StatusBarBackground),
		formatAlignedLabelValue("palette.status_bar_secondary_foreground", model.palette.StatusBarSecondaryForeground),
		formatAlignedLabelValue("palette.mode_normal_foreground", model.palette.ModeNormalForeground),
		formatAlignedLabelValue("palette.mode_normal_background", model.palette.ModeNormalBackground),
		formatAlignedLabelValue("palette.mode_insert_foreground", model.palette.ModeInsertForeground),
		formatAlignedLabelValue("palette.mode_insert_background", model.palette.ModeInsertBackground),
		formatAlignedLabelValue("run_status_icons.enabled", fmt.Sprintf("%t", model.runStatusIcons.Enabled)),
		formatAlignedLabelValue("run_status_icons.display_mode", configuredValue(model.runStatusIcons.DisplayMode)),
		formatAlignedLabelValue("run_status_icons.default_icon", configuredValue(model.runStatusIcons.DefaultIcon)),
		formatAlignedLabelValue("run_status_icons.map.succeeded", configuredValue(model.runStatusIcons.Map["succeeded"])),
		formatAlignedLabelValue("run_status_icons.map.failed", configuredValue(model.runStatusIcons.Map["failed"])),
		formatAlignedLabelValue("run_status_icons.map.skipped", configuredValue(model.runStatusIcons.Map["skipped"])),
		formatAlignedLabelValue("run_status_icons.map.pending", configuredValue(model.runStatusIcons.Map["pending"])),
		formatAlignedLabelValue("run_status_icons.map.canceled", configuredValue(model.runStatusIcons.Map["canceled"])),
		formatAlignedLabelValue("run_status_icons.map.cancelled", configuredValue(model.runStatusIcons.Map["cancelled"])),
		formatAlignedLabelValue("run_status_icons.map.inprogress", configuredValue(model.runStatusIcons.Map["inprogress"])),
		formatAlignedLabelValue("run_status_icons.map.partiallysucceeded", configuredValue(model.runStatusIcons.Map["partiallysucceeded"])),
		formatAlignedLabelValue("run_status_icons.map.notstarted", configuredValue(model.runStatusIcons.Map["notstarted"])),
		formatAlignedLabelValue("run_status_icons.map.na", configuredValue(model.runStatusIcons.Map["na"])),
		formatAlignedLabelValue("run_status_colors.enabled", fmt.Sprintf("%t", model.runStatusColors.Enabled)),
		formatAlignedLabelValue("run_status_colors.default_color", configuredValue(model.runStatusColors.DefaultColor)),
		formatAlignedLabelValue("run_status_colors.map.succeeded", configuredValue(model.runStatusColors.Map["succeeded"])),
		formatAlignedLabelValue("run_status_colors.map.failed", configuredValue(model.runStatusColors.Map["failed"])),
		formatAlignedLabelValue("run_status_colors.map.skipped", configuredValue(model.runStatusColors.Map["skipped"])),
		formatAlignedLabelValue("run_status_colors.map.pending", configuredValue(model.runStatusColors.Map["pending"])),
		formatAlignedLabelValue("run_status_colors.map.canceled", configuredValue(model.runStatusColors.Map["canceled"])),
		formatAlignedLabelValue("run_status_colors.map.cancelled", configuredValue(model.runStatusColors.Map["cancelled"])),
		formatAlignedLabelValue("run_status_colors.map.inprogress", configuredValue(model.runStatusColors.Map["inprogress"])),
		formatAlignedLabelValue("run_status_colors.map.partiallysucceeded", configuredValue(model.runStatusColors.Map["partiallysucceeded"])),
		formatAlignedLabelValue("run_status_colors.map.notstarted", configuredValue(model.runStatusColors.Map["notstarted"])),
		formatAlignedLabelValue("run_status_colors.map.na", configuredValue(model.runStatusColors.Map["na"])),
	}

	helpLines := []string{
		titleStyle.Render("Help / Keymaps"),
		"",
		sectionStyle.Render("Keymaps"),
		formatAlignedActionKeys("search", "s", "i", "/"),
		formatAlignedActionKeys("move up", "k", "up"),
		formatAlignedActionKeys("move down", "j", "down"),
		formatAlignedActionKeys("select", "enter", "l"),
		formatAlignedActionKeys("back", "esc", "b"),
		formatAlignedActionKeys("refresh", "r"),
		formatAlignedActionKeys("runs result", "alt+f"),
		formatAlignedActionKeys("quit", ":q"),
		formatAlignedActionKeys("keymap", "?"),
		formatAlignedActionKeys("command", ":"),
		formatAlignedActionKeys("run details focus", "tab", "left", "right", "h", "l", "0", "2"),
		formatAlignedActionKeys("log content nav", "j", "k", "ctrl+d", "ctrl+u", "gg", "G"),
		mutedStyle.Render("runs filter: result:<value> (example: result:failed api)"),
		mutedStyle.Render("runs API result values: " + strings.Join(allowedRunResultFilters, ", ")),
		"",
		sectionStyle.Render("Configuration"),
		formatAlignedLabelValue("config file", configuredValue(model.configFilePath)),
	}

	helpLines = append(helpLines, configLines...)
	helpLines = append(helpLines, "", mutedStyle.Render("Press ? or esc to close"))

	visibleLineCount := panelHeight - 2
	if visibleLineCount < 1 {
		visibleLineCount = 1
	}

	renderedLines := make([]string, 0, visibleLineCount)
	for index := 0; index < visibleLineCount && index < len(helpLines); index++ {
		renderedLines = append(renderedLines, helpLines[index])
	}

	return panelStyle.Render(strings.Join(renderedLines, "\n"))
}

func (model MainLayoutModel) stageTitle() string {
	switch model.currentStage {
	case stageProjects:
		return "Projects"
	case stagePipelines:
		return "Pipelines"
	case stageRuns:
		return fmt.Sprintf("Runs [result: %s]", model.activeRunResultFilterLabel())
	case stageRunDetails:
		return "Run Details"
	default:
		return "Items"
	}
}

func (model MainLayoutModel) stageContentLines(contentWidth int, contentLineBudget int) []string {
	if model.isLoading {
		loadingText := strings.TrimSpace(model.loadingMessage)
		if loadingText == "" {
			loadingText = "Loading..."
		}
		return []string{loadingText}
	}

	switch model.currentStage {
	case stageProjects:
		if len(model.projects) == 0 {
			return []string{"No projects found."}
		}
		return model.renderProjectLines(contentWidth, contentLineBudget)
	case stagePipelines:
		if len(model.pipelines) == 0 {
			return []string{"No pipelines found for selected project.", "Press esc to go back and select another project."}
		}
		return model.renderPipelineLines(contentWidth, contentLineBudget)
	case stageRuns:
		if len(model.runs) == 0 {
			return []string{
				fmt.Sprintf("No runs found for selected pipeline (result: %s).", model.activeRunResultFilterLabel()),
				"Press alt+f to select a result, or esc to go back and select another pipeline.",
			}
		}
		return model.renderRunLines(contentWidth, contentLineBudget)
	case stageRunDetails:
		return model.renderRunDetailsSplitLines(contentWidth, contentLineBudget)
	default:
		return []string{"No data."}
	}
}
