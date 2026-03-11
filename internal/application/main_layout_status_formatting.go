package application

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func normalizeStatusKey(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func compressStatusKey(status string) string {
	replacer := strings.NewReplacer(" ", "", "_", "", "-", "", ".", "", "/", "", ":", "")
	return replacer.Replace(normalizeStatusKey(status))
}

func (model MainLayoutModel) statusDisplay(rawStatus string) string {
	statusValue := normalizeStatusKey(rawStatus)
	if statusValue == "" {
		statusValue = "unknown"
	}
	displayMode := model.normalizedRunStatusDisplayMode()
	if displayMode == "text_only" || !model.runStatusIcons.Enabled {
		return statusValue
	}

	icon := model.statusIcon(statusValue)
	if icon == "" {
		if displayMode == "icons_only" {
			return ""
		}
		return statusValue
	}

	if displayMode == "icons_only" {
		return icon
	}
	return fmt.Sprintf("%s %s", icon, statusValue)
}

func (model MainLayoutModel) statusIcon(rawStatus string) string {
	statusValue := normalizeStatusKey(rawStatus)
	if statusValue == "" {
		statusValue = "unknown"
	}
	candidateKeys := statusCandidateKeys(statusValue)
	icon := ""
	for _, candidateKey := range candidateKeys {
		if model.runStatusIcons.Map == nil {
			break
		}
		if mappedIcon, ok := model.runStatusIcons.Map[candidateKey]; ok && strings.TrimSpace(mappedIcon) != "" {
			icon = strings.TrimSpace(mappedIcon)
			break
		}
	}
	if icon == "" {
		icon = strings.TrimSpace(model.runStatusIcons.DefaultIcon)
	}
	if icon == "" {
		icon = "•"
	}

	if model.runStatusColors.Enabled {
		color := ""
		for _, candidateKey := range candidateKeys {
			if model.runStatusColors.Map == nil {
				break
			}
			if mappedColor, ok := model.runStatusColors.Map[candidateKey]; ok && strings.TrimSpace(mappedColor) != "" {
				color = strings.TrimSpace(mappedColor)
				break
			}
		}
		if color == "" {
			color = strings.TrimSpace(model.runStatusColors.DefaultColor)
		}
		if color != "" {
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(icon)
		}
	}
	return icon
}

func (model MainLayoutModel) normalizedRunStatusDisplayMode() string {
	mode := strings.ToLower(strings.TrimSpace(model.runStatusIcons.DisplayMode))
	switch mode {
	case "icons_only", "icons_and_text", "text_only":
		return mode
	default:
		return "icons_and_text"
	}
}

func statusCandidateKeys(statusValue string) []string {
	candidates := make([]string, 0, 8)
	seen := map[string]bool{}
	appendCandidate := func(value string) {
		normalizedValue := normalizeStatusKey(value)
		if normalizedValue == "" || seen[normalizedValue] {
			return
		}
		seen[normalizedValue] = true
		candidates = append(candidates, normalizedValue)
		compressedValue := compressStatusKey(normalizedValue)
		if compressedValue != "" && !seen[compressedValue] {
			seen[compressedValue] = true
			candidates = append(candidates, compressedValue)
		}
	}

	appendCandidate(statusValue)
	for _, part := range strings.Split(statusValue, "/") {
		appendCandidate(part)
	}
	for _, part := range strings.Split(statusValue, "|") {
		appendCandidate(part)
	}
	if isNAStatusToken(statusValue) {
		appendCandidate("n/a")
		appendCandidate("na")
	}
	return candidates
}

func isNAStatusToken(value string) bool {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	if normalizedValue == "" {
		return false
	}
	replacer := strings.NewReplacer("(", "", ")", "", "[", "", "]", "", " ", "")
	normalizedValue = replacer.Replace(normalizedValue)
	return normalizedValue == "n/a" || normalizedValue == "na"
}

func (model MainLayoutModel) statusPrefix(rawStatus string, name string) string {
	statusText := strings.TrimSpace(model.statusDisplay(rawStatus))
	trimmedName := strings.TrimSpace(name)
	if statusText == "" {
		return trimmedName
	}
	if trimmedName == "" {
		return statusText
	}
	return statusText + " " + trimmedName
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue != "" {
			return trimmedValue
		}
	}
	return ""
}

func configuredValue(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "(empty)"
	}
	return trimmedValue
}

func formatAlignedLabelValue(label string, value string) string {
	return fmt.Sprintf("%-34s %s", label+":", value)
}

func formatAlignedActionKeys(action string, keys ...string) string {
	return fmt.Sprintf("%-14s %s", action+":", strings.Join(keys, " | "))
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
