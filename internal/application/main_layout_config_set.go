package application

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"lazydevops/internal/application/components/search"
	"lazydevops/internal/devops"
	"lazydevops/internal/settings"
)

func (model MainLayoutModel) buildSetOptions() []setOption {
	return []setOption{
		{Key: "organization", CurrentValue: configuredValue(model.organizationURL)},
		{Key: "main_layout.default_input_mode", CurrentValue: model.mainLayoutSettings.DefaultInputMode},
		{Key: "main_layout.show_path_in_title", CurrentValue: fmt.Sprintf("%t", model.mainLayoutSettings.ShowPathInTitle)},
		{Key: "main_layout.show_path_in_status_bar", CurrentValue: fmt.Sprintf("%t", model.mainLayoutSettings.ShowPathInStatusBar)},
		{Key: "main_layout.status_bar_path_side", CurrentValue: model.mainLayoutSettings.StatusBarPathSide},
		{Key: "main_layout.list_highlight_mode", CurrentValue: configuredValue(model.mainLayoutSettings.ListHighlightMode)},
		{Key: "main_layout.default_project", CurrentValue: configuredValue(model.mainLayoutSettings.DefaultProject)},
		{Key: "search.match_highlight_mode", CurrentValue: configuredValue(model.searchSettings.MatchHighlightMode)},
		{Key: "search.rainbow_colors", CurrentValue: configuredValue(strings.Join(model.searchSettings.RainbowColors, ","))},
		{Key: "log_rendering.mode", CurrentValue: configuredValue(model.logRenderingSettings.Mode)},
		{Key: "palette.title", CurrentValue: model.palette.Title},
		{Key: "palette.info", CurrentValue: model.palette.Info},
		{Key: "palette.error", CurrentValue: model.palette.Error},
		{Key: "palette.help", CurrentValue: model.palette.Help},
		{Key: "palette.accent", CurrentValue: model.palette.Accent},
		{Key: "palette.spinner", CurrentValue: model.palette.Spinner},
		{Key: "palette.success", CurrentValue: model.palette.Success},
		{Key: "palette.muted", CurrentValue: model.palette.Muted},
		{Key: "palette.selected_foreground", CurrentValue: model.palette.SelectedForeground},
		{Key: "palette.selected_background", CurrentValue: model.palette.SelectedBackground},
		{Key: "palette.insert_selected_foreground", CurrentValue: model.palette.InsertSelectedForeground},
		{Key: "palette.insert_selected_background", CurrentValue: model.palette.InsertSelectedBackground},
		{Key: "palette.status_bar_foreground", CurrentValue: model.palette.StatusBarForeground},
		{Key: "palette.status_bar_background", CurrentValue: model.palette.StatusBarBackground},
		{Key: "palette.status_bar_secondary_foreground", CurrentValue: model.palette.StatusBarSecondaryForeground},
		{Key: "palette.mode_normal_foreground", CurrentValue: model.palette.ModeNormalForeground},
		{Key: "palette.mode_normal_background", CurrentValue: model.palette.ModeNormalBackground},
		{Key: "palette.mode_insert_foreground", CurrentValue: model.palette.ModeInsertForeground},
		{Key: "palette.mode_insert_background", CurrentValue: model.palette.ModeInsertBackground},
		{Key: "run_status_icons.enabled", CurrentValue: fmt.Sprintf("%t", model.runStatusIcons.Enabled)},
		{Key: "run_status_icons.display_mode", CurrentValue: model.runStatusIcons.DisplayMode},
		{Key: "run_status_icons.default_icon", CurrentValue: model.runStatusIcons.DefaultIcon},
		{Key: "run_status_icons.map.succeeded", CurrentValue: model.runStatusIcons.Map["succeeded"]},
		{Key: "run_status_icons.map.failed", CurrentValue: model.runStatusIcons.Map["failed"]},
		{Key: "run_status_icons.map.skipped", CurrentValue: model.runStatusIcons.Map["skipped"]},
		{Key: "run_status_icons.map.pending", CurrentValue: model.runStatusIcons.Map["pending"]},
		{Key: "run_status_icons.map.canceled", CurrentValue: model.runStatusIcons.Map["canceled"]},
		{Key: "run_status_icons.map.cancelled", CurrentValue: model.runStatusIcons.Map["cancelled"]},
		{Key: "run_status_icons.map.inprogress", CurrentValue: model.runStatusIcons.Map["inprogress"]},
		{Key: "run_status_icons.map.partiallysucceeded", CurrentValue: model.runStatusIcons.Map["partiallysucceeded"]},
		{Key: "run_status_icons.map.notstarted", CurrentValue: model.runStatusIcons.Map["notstarted"]},
		{Key: "run_status_icons.map.na", CurrentValue: model.runStatusIcons.Map["na"]},
		{Key: "run_status_colors.enabled", CurrentValue: fmt.Sprintf("%t", model.runStatusColors.Enabled)},
		{Key: "run_status_colors.default_color", CurrentValue: model.runStatusColors.DefaultColor},
		{Key: "run_status_colors.map.succeeded", CurrentValue: model.runStatusColors.Map["succeeded"]},
		{Key: "run_status_colors.map.failed", CurrentValue: model.runStatusColors.Map["failed"]},
		{Key: "run_status_colors.map.skipped", CurrentValue: model.runStatusColors.Map["skipped"]},
		{Key: "run_status_colors.map.pending", CurrentValue: model.runStatusColors.Map["pending"]},
		{Key: "run_status_colors.map.canceled", CurrentValue: model.runStatusColors.Map["canceled"]},
		{Key: "run_status_colors.map.cancelled", CurrentValue: model.runStatusColors.Map["cancelled"]},
		{Key: "run_status_colors.map.inprogress", CurrentValue: model.runStatusColors.Map["inprogress"]},
		{Key: "run_status_colors.map.partiallysucceeded", CurrentValue: model.runStatusColors.Map["partiallysucceeded"]},
		{Key: "run_status_colors.map.notstarted", CurrentValue: model.runStatusColors.Map["notstarted"]},
		{Key: "run_status_colors.map.na", CurrentValue: model.runStatusColors.Map["na"]},
	}
}

func (model MainLayoutModel) filteredSetOptions() []setOption {
	allOptions := model.setOptions
	if len(allOptions) == 0 {
		return nil
	}
	filterValue := strings.ToLower(strings.TrimSpace(model.setSearchInput.Value()))
	if filterValue == "" {
		return allOptions
	}
	filteredOptions := make([]setOption, 0, len(allOptions))
	for _, option := range allOptions {
		if search.Matches(option.Key+" "+option.CurrentValue, filterValue) {
			filteredOptions = append(filteredOptions, option)
		}
	}
	return filteredOptions
}

func validateSetValue(key string, value string) error {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))

	switch key {
	case "organization":
		if value == "" {
			return fmt.Errorf("organization cannot be empty")
		}
		organizationPattern := regexp.MustCompile(`^https://dev\\.azure\\.com/[A-Za-z0-9._-]+/?$`)
		if !organizationPattern.MatchString(value) {
			return fmt.Errorf("organization must match https://dev.azure.com/<org>")
		}
		return nil
	case "main_layout.default_input_mode":
		if normalizedValue != "normal" && normalizedValue != "insert" {
			return fmt.Errorf("default_input_mode must be normal or insert")
		}
		return nil
	case "main_layout.show_path_in_title", "main_layout.show_path_in_status_bar", "run_status_icons.enabled", "run_status_colors.enabled":
		if normalizedValue != "true" && normalizedValue != "false" {
			return fmt.Errorf("value must be true or false")
		}
		return nil
	case "run_status_icons.display_mode":
		if normalizedValue != "icons_only" && normalizedValue != "icons_and_text" && normalizedValue != "text_only" {
			return fmt.Errorf("display_mode must be icons_only, icons_and_text, or text_only")
		}
		return nil
	case "main_layout.status_bar_path_side":
		if normalizedValue != "left" && normalizedValue != "right" {
			return fmt.Errorf("status_bar_path_side must be left or right")
		}
		return nil
	case "main_layout.list_highlight_mode":
		if normalizedValue != "off" && normalizedValue != "accent" && normalizedValue != "selection" {
			return fmt.Errorf("list_highlight_mode must be off, accent, or selection")
		}
		return nil
	case "main_layout.default_project":
		return nil
	case "search.match_highlight_mode":
		if normalizedValue != "accent" && normalizedValue != "rainbow" {
			return fmt.Errorf("match_highlight_mode must be accent or rainbow")
		}
		return nil
	case "log_rendering.mode":
		if normalizedValue != "auto" && normalizedValue != "plain" {
			return fmt.Errorf("log_rendering.mode must be auto or plain")
		}
		return nil
	case "search.rainbow_colors":
		if value == "" {
			return fmt.Errorf("rainbow_colors cannot be empty")
		}
		hexPattern := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
		colorValues := strings.Split(value, ",")
		validColorCount := 0
		for _, colorValue := range colorValues {
			trimmedColor := strings.TrimSpace(colorValue)
			if trimmedColor == "" {
				continue
			}
			if !hexPattern.MatchString(trimmedColor) {
				return fmt.Errorf("rainbow_colors must be comma-separated hex colors")
			}
			validColorCount++
		}
		if validColorCount == 0 {
			return fmt.Errorf("rainbow_colors must include at least one hex color")
		}
		return nil
	case "run_status_icons.default_icon":
		if value == "" {
			return fmt.Errorf("default_icon cannot be empty")
		}
		return nil
	case "run_status_colors.default_color":
		if value == "" {
			return fmt.Errorf("default_color cannot be empty")
		}
		hexPattern := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
		if !hexPattern.MatchString(value) {
			return fmt.Errorf("status icon colors must be hex (#RGB, #RRGGBB, or #RRGGBBAA)")
		}
		return nil
	}

	if strings.HasPrefix(key, "palette.") {
		hexPattern := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
		if !hexPattern.MatchString(value) {
			return fmt.Errorf("palette colors must be hex (#RGB, #RRGGBB, or #RRGGBBAA)")
		}
		return nil
	}
	if strings.HasPrefix(key, "run_status_icons.map.") {
		if value == "" {
			return fmt.Errorf("status icon cannot be empty")
		}
		return nil
	}
	if strings.HasPrefix(key, "run_status_colors.map.") {
		hexPattern := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
		if !hexPattern.MatchString(value) {
			return fmt.Errorf("status icon colors must be hex (#RGB, #RRGGBB, or #RRGGBBAA)")
		}
		return nil
	}

	return fmt.Errorf("unsupported setting key")
}

func setKeyAllowedValues(key string) []string {
	switch key {
	case "main_layout.default_input_mode":
		return []string{"normal", "insert"}
	case "main_layout.show_path_in_title", "main_layout.show_path_in_status_bar", "run_status_icons.enabled", "run_status_colors.enabled":
		return []string{"true", "false"}
	case "main_layout.status_bar_path_side":
		return []string{"left", "right"}
	case "main_layout.list_highlight_mode":
		return []string{"off", "accent", "selection"}
	case "search.match_highlight_mode":
		return []string{"accent", "rainbow"}
	case "log_rendering.mode":
		return []string{"auto", "plain"}
	case "run_status_icons.display_mode":
		return []string{"icons_only", "icons_and_text", "text_only"}
	}
	return nil
}

func (model MainLayoutModel) applySetValue() (MainLayoutModel, error) {
	value := strings.TrimSpace(model.setPendingValue)
	normalizedValue := strings.ToLower(value)
	switch model.setSelectedOption.Key {
	case "organization":
		model.organizationURL = value
		model.devopsService = devops.NewService(model.organizationURL)
	case "main_layout.default_input_mode":
		model.mainLayoutSettings.DefaultInputMode = normalizedValue
	case "main_layout.show_path_in_title":
		model.mainLayoutSettings.ShowPathInTitle = normalizedValue == "true"
	case "main_layout.show_path_in_status_bar":
		model.mainLayoutSettings.ShowPathInStatusBar = normalizedValue == "true"
	case "main_layout.status_bar_path_side":
		model.mainLayoutSettings.StatusBarPathSide = normalizedValue
	case "main_layout.list_highlight_mode":
		model.mainLayoutSettings.ListHighlightMode = normalizedValue
	case "main_layout.default_project":
		model.mainLayoutSettings.DefaultProject = value
	case "search.match_highlight_mode":
		model.searchSettings.MatchHighlightMode = normalizedValue
	case "log_rendering.mode":
		model.logRenderingSettings.Mode = normalizedValue
	case "search.rainbow_colors":
		colors := strings.Split(value, ",")
		updatedColors := make([]string, 0, len(colors))
		for _, color := range colors {
			trimmedColor := strings.TrimSpace(color)
			if trimmedColor != "" {
				updatedColors = append(updatedColors, trimmedColor)
			}
		}
		if len(updatedColors) > 0 {
			model.searchSettings.RainbowColors = updatedColors
		}
	case "palette.title":
		model.palette.Title = value
	case "palette.info":
		model.palette.Info = value
	case "palette.error":
		model.palette.Error = value
	case "palette.help":
		model.palette.Help = value
	case "palette.accent":
		model.palette.Accent = value
	case "palette.spinner":
		model.palette.Spinner = value
		model.spinnerModel.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Spinner))
	case "palette.success":
		model.palette.Success = value
	case "palette.muted":
		model.palette.Muted = value
	case "palette.selected_foreground":
		model.palette.SelectedForeground = value
	case "palette.selected_background":
		model.palette.SelectedBackground = value
	case "palette.insert_selected_foreground":
		model.palette.InsertSelectedForeground = value
	case "palette.insert_selected_background":
		model.palette.InsertSelectedBackground = value
	case "palette.status_bar_foreground":
		model.palette.StatusBarForeground = value
	case "palette.status_bar_background":
		model.palette.StatusBarBackground = value
	case "palette.status_bar_secondary_foreground":
		model.palette.StatusBarSecondaryForeground = value
	case "palette.mode_normal_foreground":
		model.palette.ModeNormalForeground = value
	case "palette.mode_normal_background":
		model.palette.ModeNormalBackground = value
	case "palette.mode_insert_foreground":
		model.palette.ModeInsertForeground = value
	case "palette.mode_insert_background":
		model.palette.ModeInsertBackground = value
	case "run_status_icons.enabled":
		model.runStatusIcons.Enabled = normalizedValue == "true"
	case "run_status_icons.display_mode":
		model.runStatusIcons.DisplayMode = normalizedValue
	case "run_status_icons.default_icon":
		model.runStatusIcons.DefaultIcon = value
	case "run_status_colors.enabled":
		model.runStatusColors.Enabled = normalizedValue == "true"
	case "run_status_colors.default_color":
		model.runStatusColors.DefaultColor = value
	default:
		if strings.HasPrefix(model.setSelectedOption.Key, "run_status_icons.map.") {
			status := strings.TrimPrefix(model.setSelectedOption.Key, "run_status_icons.map.")
			if model.runStatusIcons.Map == nil {
				model.runStatusIcons.Map = map[string]string{}
			}
			model.runStatusIcons.Map[normalizeStatusKey(status)] = value
		} else if strings.HasPrefix(model.setSelectedOption.Key, "run_status_colors.map.") {
			status := strings.TrimPrefix(model.setSelectedOption.Key, "run_status_colors.map.")
			if model.runStatusColors.Map == nil {
				model.runStatusColors.Map = map[string]string{}
			}
			model.runStatusColors.Map[normalizeStatusKey(status)] = value
		} else {
			return model, fmt.Errorf("unsupported setting key: %s", model.setSelectedOption.Key)
		}
	}

	updatedSettings := settings.Settings{
		Organization:    model.organizationURL,
		MainLayout:      model.mainLayoutSettings,
		Search:          model.searchSettings,
		LogRendering:    model.logRenderingSettings,
		Palette:         model.palette,
		RunStatusIcons:  model.runStatusIcons,
		RunStatusColors: model.runStatusColors,
	}
	if saveError := settings.Save(updatedSettings); saveError != nil {
		return model, fmt.Errorf("failed to save config: %w", saveError)
	}
	reloadedSettings := settings.Reload()
	model.mainLayoutSettings = reloadedSettings.MainLayout
	model.searchSettings = reloadedSettings.Search
	model.logRenderingSettings = reloadedSettings.LogRendering
	model.palette = reloadedSettings.Palette
	model.runStatusIcons = reloadedSettings.RunStatusIcons
	model.runStatusColors = reloadedSettings.RunStatusColors
	if strings.TrimSpace(reloadedSettings.Organization) != "" {
		model.organizationURL = strings.TrimSpace(reloadedSettings.Organization)
		model.devopsService = devops.NewService(model.organizationURL)
	}
	model.spinnerModel.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(model.palette.Spinner))

	return model, nil
}
