package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Settings struct {
	Organization    string                 `json:"organization"`
	MainLayout      MainLayoutSettings     `json:"main_layout"`
	Search          SearchSettings         `json:"search"`
	LogRendering    LogRenderingSettings   `json:"log_rendering"`
	Palette         Palette                `json:"palette"`
	RunStatusIcons  RunStatusIconSettings  `json:"run_status_icons"`
	RunStatusColors RunStatusColorSettings `json:"run_status_colors"`
}

type configuredSettings struct {
	Organization    string                           `json:"organization"`
	MainLayout      configuredMainLayoutSettings     `json:"main_layout"`
	Search          configuredSearchSettings         `json:"search"`
	LogRendering    configuredLogRenderingSettings   `json:"log_rendering"`
	Palette         Palette                          `json:"palette"`
	RunStatusIcons  configuredRunStatusIconSettings  `json:"run_status_icons"`
	RunStatusColors configuredRunStatusColorSettings `json:"run_status_colors"`
}

type configuredMainLayoutSettings struct {
	DefaultInputMode    string `json:"default_input_mode"`
	StatusBarPathSide   string `json:"status_bar_path_side"`
	DefaultProject      string `json:"default_project"`
	ListHighlightMode   string `json:"list_highlight_mode"`
	ShowPathInTitle     *bool  `json:"show_path_in_title"`
	ShowPathInStatusBar *bool  `json:"show_path_in_status_bar"`
}

type configuredRunStatusIconSettings struct {
	Enabled     *bool             `json:"enabled"`
	DisplayMode string            `json:"display_mode"`
	DefaultIcon string            `json:"default_icon"`
	Map         map[string]string `json:"map"`
}

type configuredRunStatusColorSettings struct {
	Enabled      *bool             `json:"enabled"`
	DefaultColor string            `json:"default_color"`
	Map          map[string]string `json:"map"`
}

type configuredSearchSettings struct {
	MatchHighlightMode string   `json:"match_highlight_mode"`
	RainbowColors      []string `json:"rainbow_colors"`
}

type configuredLogRenderingSettings struct {
	Mode string `json:"mode"`
}

type MainLayoutSettings struct {
	DefaultInputMode    string `json:"default_input_mode"`
	StatusBarPathSide   string `json:"status_bar_path_side"`
	DefaultProject      string `json:"default_project"`
	ListHighlightMode   string `json:"list_highlight_mode"`
	ShowPathInTitle     bool   `json:"show_path_in_title"`
	ShowPathInStatusBar bool   `json:"show_path_in_status_bar"`
}

type RunStatusIconSettings struct {
	Enabled     bool              `json:"enabled"`
	DisplayMode string            `json:"display_mode"`
	DefaultIcon string            `json:"default_icon"`
	Map         map[string]string `json:"map"`
}

type RunStatusColorSettings struct {
	Enabled      bool              `json:"enabled"`
	DefaultColor string            `json:"default_color"`
	Map          map[string]string `json:"map"`
}

type SearchSettings struct {
	MatchHighlightMode string   `json:"match_highlight_mode"`
	RainbowColors      []string `json:"rainbow_colors"`
}

type LogRenderingSettings struct {
	Mode string `json:"mode"`
}

type Palette struct {
	Title                        string `json:"title"`
	Info                         string `json:"info"`
	Error                        string `json:"error"`
	Help                         string `json:"help"`
	Accent                       string `json:"accent"`
	Spinner                      string `json:"spinner"`
	Success                      string `json:"success"`
	Muted                        string `json:"muted"`
	SelectedForeground           string `json:"selected_foreground"`
	SelectedBackground           string `json:"selected_background"`
	InsertSelectedForeground     string `json:"insert_selected_foreground"`
	InsertSelectedBackground     string `json:"insert_selected_background"`
	StatusBarForeground          string `json:"status_bar_foreground"`
	StatusBarBackground          string `json:"status_bar_background"`
	StatusBarSecondaryForeground string `json:"status_bar_secondary_foreground"`
	ModeNormalForeground         string `json:"mode_normal_foreground"`
	ModeNormalBackground         string `json:"mode_normal_background"`
	ModeInsertForeground         string `json:"mode_insert_foreground"`
	ModeInsertBackground         string `json:"mode_insert_background"`
}

var (
	settingsMutex  sync.Mutex
	isLoaded       bool
	cachedSettings Settings
)

func Current() Settings {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	if !isLoaded {
		cachedSettings = loadFromDisk()
		isLoaded = true
	}

	return cachedSettings
}

func Reload() Settings {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	cachedSettings = loadFromDisk()
	isLoaded = true
	return cachedSettings
}

func Save(updatedSettings Settings) error {
	configPath, pathError := ConfigFilePath()
	if pathError != nil {
		return pathError
	}

	configDirectory := filepath.Dir(configPath)
	if mkdirError := os.MkdirAll(configDirectory, 0o755); mkdirError != nil {
		return mkdirError
	}

	encodedSettings, encodeError := json.MarshalIndent(updatedSettings, "", "  ")
	if encodeError != nil {
		return encodeError
	}
	encodedSettings = append(encodedSettings, '\n')

	if writeError := os.WriteFile(configPath, encodedSettings, 0o644); writeError != nil {
		return writeError
	}

	settingsMutex.Lock()
	cachedSettings = updatedSettings
	isLoaded = true
	settingsMutex.Unlock()

	return nil
}

func ConfigFilePath() (string, error) {
	userConfigDirectory, resolveError := os.UserConfigDir()
	if resolveError != nil {
		return "", resolveError
	}

	return filepath.Join(userConfigDirectory, "lazydevops", "config.json"), nil
}

func loadFromDisk() Settings {
	loaded := defaultSettings()
	configPath, pathError := ConfigFilePath()
	if pathError != nil {
		return loaded
	}

	settingsBytes, readError := os.ReadFile(configPath)
	if readError != nil {
		return loaded
	}

	var configuredValues configuredSettings
	if decodeError := json.Unmarshal(settingsBytes, &configuredValues); decodeError != nil {
		return loaded
	}

	return mergeSettings(loaded, configuredValues)
}

func mergeSettings(defaultValues Settings, configuredValues configuredSettings) Settings {
	merged := defaultValues

	if strings.TrimSpace(configuredValues.Organization) != "" {
		merged.Organization = strings.TrimSpace(configuredValues.Organization)
	}
	if isValidInputMode(configuredValues.MainLayout.DefaultInputMode) {
		merged.MainLayout.DefaultInputMode = strings.ToLower(strings.TrimSpace(configuredValues.MainLayout.DefaultInputMode))
	}
	if isValidStatusBarPathSide(configuredValues.MainLayout.StatusBarPathSide) {
		merged.MainLayout.StatusBarPathSide = strings.ToLower(strings.TrimSpace(configuredValues.MainLayout.StatusBarPathSide))
	}
	if isValidListHighlightMode(configuredValues.MainLayout.ListHighlightMode) {
		merged.MainLayout.ListHighlightMode = normalizeListHighlightMode(configuredValues.MainLayout.ListHighlightMode)
	}
	if strings.TrimSpace(configuredValues.MainLayout.DefaultProject) != "" {
		merged.MainLayout.DefaultProject = strings.TrimSpace(configuredValues.MainLayout.DefaultProject)
	}
	if configuredValues.MainLayout.ShowPathInTitle != nil {
		merged.MainLayout.ShowPathInTitle = *configuredValues.MainLayout.ShowPathInTitle
	}
	if configuredValues.MainLayout.ShowPathInStatusBar != nil {
		merged.MainLayout.ShowPathInStatusBar = *configuredValues.MainLayout.ShowPathInStatusBar
	}

	merged.Search = mergeSearchSettings(defaultValues.Search, configuredValues.Search)
	merged.LogRendering = mergeLogRenderingSettings(defaultValues.LogRendering, configuredValues.LogRendering)
	merged.Palette = mergePalette(defaultValues.Palette, configuredValues.Palette)
	merged.RunStatusIcons = mergeRunStatusIcons(defaultValues.RunStatusIcons, configuredValues.RunStatusIcons)
	merged.RunStatusColors = mergeRunStatusColors(defaultValues.RunStatusColors, configuredValues.RunStatusColors)
	return merged
}

func mergeSearchSettings(defaultSearch SearchSettings, configuredSearch configuredSearchSettings) SearchSettings {
	merged := defaultSearch

	if isValidMatchHighlightMode(configuredSearch.MatchHighlightMode) {
		merged.MatchHighlightMode = normalizeMatchHighlightMode(configuredSearch.MatchHighlightMode)
	}
	if len(configuredSearch.RainbowColors) > 0 {
		filtered := make([]string, 0, len(configuredSearch.RainbowColors))
		for _, color := range configuredSearch.RainbowColors {
			trimmed := strings.TrimSpace(color)
			if trimmed == "" {
				continue
			}
			filtered = append(filtered, trimmed)
		}
		if len(filtered) > 0 {
			merged.RainbowColors = filtered
		}
	}

	return merged
}

func mergeLogRenderingSettings(defaultLogRendering LogRenderingSettings, configuredLogRendering configuredLogRenderingSettings) LogRenderingSettings {
	merged := defaultLogRendering
	if isValidLogRenderingMode(configuredLogRendering.Mode) {
		merged.Mode = normalizeLogRenderingMode(configuredLogRendering.Mode)
	}
	return merged
}

func mergePalette(defaultPalette Palette, configuredPalette Palette) Palette {
	merged := defaultPalette

	if configuredPalette.Title != "" {
		merged.Title = configuredPalette.Title
	}
	if configuredPalette.Info != "" {
		merged.Info = configuredPalette.Info
	}
	if configuredPalette.Error != "" {
		merged.Error = configuredPalette.Error
	}
	if configuredPalette.Help != "" {
		merged.Help = configuredPalette.Help
	}
	if configuredPalette.Accent != "" {
		merged.Accent = configuredPalette.Accent
	}
	if configuredPalette.Spinner != "" {
		merged.Spinner = configuredPalette.Spinner
	}
	if configuredPalette.Success != "" {
		merged.Success = configuredPalette.Success
	}
	if configuredPalette.Muted != "" {
		merged.Muted = configuredPalette.Muted
	}
	if configuredPalette.SelectedForeground != "" {
		merged.SelectedForeground = configuredPalette.SelectedForeground
	}
	if configuredPalette.SelectedBackground != "" {
		merged.SelectedBackground = configuredPalette.SelectedBackground
	}
	if configuredPalette.InsertSelectedForeground != "" {
		merged.InsertSelectedForeground = configuredPalette.InsertSelectedForeground
	}
	if configuredPalette.InsertSelectedBackground != "" {
		merged.InsertSelectedBackground = configuredPalette.InsertSelectedBackground
	}
	if configuredPalette.StatusBarForeground != "" {
		merged.StatusBarForeground = configuredPalette.StatusBarForeground
	}
	if configuredPalette.StatusBarBackground != "" {
		merged.StatusBarBackground = configuredPalette.StatusBarBackground
	}
	if configuredPalette.StatusBarSecondaryForeground != "" {
		merged.StatusBarSecondaryForeground = configuredPalette.StatusBarSecondaryForeground
	}
	if configuredPalette.ModeNormalForeground != "" {
		merged.ModeNormalForeground = configuredPalette.ModeNormalForeground
	}
	if configuredPalette.ModeNormalBackground != "" {
		merged.ModeNormalBackground = configuredPalette.ModeNormalBackground
	}
	if configuredPalette.ModeInsertForeground != "" {
		merged.ModeInsertForeground = configuredPalette.ModeInsertForeground
	}
	if configuredPalette.ModeInsertBackground != "" {
		merged.ModeInsertBackground = configuredPalette.ModeInsertBackground
	}

	return merged
}

func mergeRunStatusIcons(defaultIcons RunStatusIconSettings, configuredIcons configuredRunStatusIconSettings) RunStatusIconSettings {
	merged := defaultIcons

	if configuredIcons.Enabled != nil {
		merged.Enabled = *configuredIcons.Enabled
	}
	if isValidRunStatusDisplayMode(configuredIcons.DisplayMode) {
		merged.DisplayMode = normalizeRunStatusDisplayMode(configuredIcons.DisplayMode)
	}
	if strings.TrimSpace(configuredIcons.DefaultIcon) != "" {
		merged.DefaultIcon = strings.TrimSpace(configuredIcons.DefaultIcon)
	}
	if configuredIcons.Map != nil {
		if merged.Map == nil {
			merged.Map = map[string]string{}
		}
		for status, icon := range configuredIcons.Map {
			normalizedStatus := normalizeStatusKey(status)
			trimmedIcon := strings.TrimSpace(icon)
			if normalizedStatus == "" || trimmedIcon == "" {
				continue
			}
			merged.Map[normalizedStatus] = trimmedIcon
		}
	}

	return merged
}

func mergeRunStatusColors(defaultColors RunStatusColorSettings, configuredColors configuredRunStatusColorSettings) RunStatusColorSettings {
	merged := defaultColors

	if configuredColors.Enabled != nil {
		merged.Enabled = *configuredColors.Enabled
	}
	if strings.TrimSpace(configuredColors.DefaultColor) != "" {
		merged.DefaultColor = strings.TrimSpace(configuredColors.DefaultColor)
	}
	if configuredColors.Map != nil {
		if merged.Map == nil {
			merged.Map = map[string]string{}
		}
		for status, color := range configuredColors.Map {
			normalizedStatus := normalizeStatusKey(status)
			trimmedColor := strings.TrimSpace(color)
			if normalizedStatus == "" || trimmedColor == "" {
				continue
			}
			merged.Map[normalizedStatus] = trimmedColor
		}
	}

	return merged
}

func normalizeStatusKey(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func normalizeRunStatusDisplayMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

func isValidRunStatusDisplayMode(mode string) bool {
	normalizedMode := normalizeRunStatusDisplayMode(mode)
	return normalizedMode == "icons_only" || normalizedMode == "icons_and_text" || normalizedMode == "text_only"
}

func isValidInputMode(mode string) bool {
	normalizedMode := strings.ToLower(strings.TrimSpace(mode))
	return normalizedMode == "normal" || normalizedMode == "insert"
}

func isValidStatusBarPathSide(side string) bool {
	normalizedSide := strings.ToLower(strings.TrimSpace(side))
	return normalizedSide == "left" || normalizedSide == "right"
}

func normalizeListHighlightMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

func isValidListHighlightMode(mode string) bool {
	normalizedMode := normalizeListHighlightMode(mode)
	return normalizedMode == "off" || normalizedMode == "accent" || normalizedMode == "selection"
}

func normalizeMatchHighlightMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

func isValidMatchHighlightMode(mode string) bool {
	normalizedMode := normalizeMatchHighlightMode(mode)
	return normalizedMode == "accent" || normalizedMode == "rainbow"
}

func normalizeLogRenderingMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

func isValidLogRenderingMode(mode string) bool {
	normalizedMode := normalizeLogRenderingMode(mode)
	return normalizedMode == "auto" || normalizedMode == "plain"
}

func defaultSettings() Settings {
	return Settings{
		Organization: "",
		MainLayout: MainLayoutSettings{
			DefaultInputMode:    "insert",
			StatusBarPathSide:   "right",
			DefaultProject:      "",
			ListHighlightMode:   "off",
			ShowPathInTitle:     true,
			ShowPathInStatusBar: false,
		},
		Search: SearchSettings{
			MatchHighlightMode: "accent",
			RainbowColors: []string{
				"#f38ba8",
				"#fab387",
				"#f9e2af",
				"#a6e3a1",
				"#89dceb",
				"#89b4fa",
				"#cba6f7",
			},
		},
		LogRendering: LogRenderingSettings{
			Mode: "auto",
		},
		Palette: Palette{
			Title:                        "212",
			Info:                         "252",
			Error:                        "196",
			Help:                         "241",
			Accent:                       "#89b4fa",
			Spinner:                      "69",
			Success:                      "42",
			Muted:                        "244",
			SelectedForeground:           "230",
			SelectedBackground:           "62",
			InsertSelectedForeground:     "252",
			InsertSelectedBackground:     "239",
			StatusBarForeground:          "255",
			StatusBarBackground:          "236",
			StatusBarSecondaryForeground: "252",
			ModeNormalForeground:         "230",
			ModeNormalBackground:         "62",
			ModeInsertForeground:         "230",
			ModeInsertBackground:         "166",
		},
		RunStatusIcons: RunStatusIconSettings{
			Enabled:     true,
			DisplayMode: "icons_and_text",
			DefaultIcon: "[?]",
			Map: map[string]string{
				"succeeded":          "[OK]",
				"failed":             "[X]",
				"skipped":            "[-]",
				"pending":            "[...]",
				"canceled":           "[!]",
				"cancelled":          "[!]",
				"inprogress":         "[~]",
				"partiallysucceeded": "[~]",
				"notstarted":         "[...]",
				"na":                 "[-]",
			},
		},
		RunStatusColors: RunStatusColorSettings{
			Enabled:      true,
			DefaultColor: "#a4a8b3",
			Map: map[string]string{
				"succeeded":          "#50fa7b",
				"failed":             "#ff5555",
				"skipped":            "#a4a8b3",
				"pending":            "#a4a8b3",
				"canceled":           "#ffb86c",
				"cancelled":          "#ffb86c",
				"inprogress":         "#8be9fd",
				"partiallysucceeded": "#f1fa8c",
				"notstarted":         "#a4a8b3",
				"na":                 "#a4a8b3",
			},
		},
	}
}
