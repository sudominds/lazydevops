package application

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
)

var (
	batResolveOnce      sync.Once
	resolvedBatExecPath string
	logErrorWordPattern = regexp.MustCompile(`(?i)\b(error|failed)\b`)
	logWarningPattern   = regexp.MustCompile(`(?i)\bwarning\b`)
)

func resolveBatExecutablePath() string {
	batResolveOnce.Do(func() {
		if batPath, lookupError := exec.LookPath("bat"); lookupError == nil {
			resolvedBatExecPath = batPath
			return
		}
		if batcatPath, lookupError := exec.LookPath("batcat"); lookupError == nil {
			resolvedBatExecPath = batcatPath
		}
	})
	return strings.TrimSpace(resolvedBatExecPath)
}

func (model MainLayoutModel) normalizedLogRenderingMode() string {
	mode := strings.ToLower(strings.TrimSpace(model.logRenderingSettings.Mode))
	switch mode {
	case "plain", "auto":
		return mode
	default:
		return "auto"
	}
}

func (model MainLayoutModel) shouldUseBatLogRendering() bool {
	if model.normalizedLogRenderingMode() == "plain" {
		return false
	}
	return resolveBatExecutablePath() != ""
}

func (model MainLayoutModel) renderRunLogContentWindowLines(allContentLines []string, startOffset int, endOffset int, contentWidth int) []string {
	if len(allContentLines) == 0 {
		return nil
	}
	if startOffset < 0 {
		startOffset = 0
	}
	if endOffset > len(allContentLines) {
		endOffset = len(allContentLines)
	}
	if endOffset < startOffset {
		endOffset = startOffset
	}

	if !model.shouldUseBatLogRendering() {
		plainLines := make([]string, 0, endOffset-startOffset)
		for _, contentLine := range allContentLines[startOffset:endOffset] {
			fittedLine := fitLineToWidth(sanitizeDisplayLine(contentLine), contentWidth)
			plainLines = append(plainLines, highlightLogSeverityKeywords(fittedLine))
		}
		return plainLines
	}

	highlightedLines, hasHighlighted := model.highlightedLogLinesForSelectedLog(allContentLines)
	if !hasHighlighted || len(highlightedLines) == 0 {
		plainLines := make([]string, 0, endOffset-startOffset)
		for _, contentLine := range allContentLines[startOffset:endOffset] {
			fittedLine := fitLineToWidth(sanitizeDisplayLine(contentLine), contentWidth)
			plainLines = append(plainLines, highlightLogSeverityKeywords(fittedLine))
		}
		return plainLines
	}
	if endOffset > len(highlightedLines) {
		endOffset = len(highlightedLines)
	}

	renderedLines := make([]string, 0, endOffset-startOffset)
	for _, highlightedLine := range highlightedLines[startOffset:endOffset] {
		fittedLine := fitANSILineToWidth(highlightedLine, contentWidth)
		renderedLines = append(renderedLines, highlightLogSeverityKeywords(fittedLine))
	}
	return renderedLines
}

func (model MainLayoutModel) highlightedLogLinesForSelectedLog(allContentLines []string) ([]string, bool) {
	selectedLogID := strings.TrimSpace(model.selectedLogID)
	if selectedLogID == "" {
		return nil, false
	}
	if model.highlightedLogContentCache == nil {
		model.highlightedLogContentCache = map[string][]string{}
	}
	if model.highlightLogFailureCache == nil {
		model.highlightLogFailureCache = map[string]bool{}
	}
	if cachedLines, ok := model.highlightedLogContentCache[selectedLogID]; ok && len(cachedLines) > 0 {
		return cachedLines, true
	}
	if model.highlightLogFailureCache[selectedLogID] {
		return nil, false
	}

	highlightedLines, highlightError := highlightLogLinesWithBat(allContentLines)
	if highlightError != nil || len(highlightedLines) == 0 {
		model.highlightLogFailureCache[selectedLogID] = true
		return nil, false
	}
	model.highlightedLogContentCache[selectedLogID] = highlightedLines
	return highlightedLines, true
}

func highlightLogLinesWithBat(contentLines []string) ([]string, error) {
	batPath := resolveBatExecutablePath()
	if batPath == "" {
		return nil, fmt.Errorf("bat executable not available")
	}
	content := strings.Join(contentLines, "\n")
	arguments := []string{"--paging=never", "--color=always", "--style=plain"}
	if looksLikeJSON(contentLines) {
		arguments = append(arguments, "--language=json")
	}
	command := exec.Command(batPath, arguments...)
	command.Stdin = strings.NewReader(content)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return nil, fmt.Errorf("bat highlight failed: %w", commandError)
	}
	highlighted := strings.TrimRight(string(commandOutput), "\n")
	if strings.TrimSpace(highlighted) == "" {
		return nil, fmt.Errorf("bat returned empty highlighted output")
	}
	return strings.Split(highlighted, "\n"), nil
}

func looksLikeJSON(contentLines []string) bool {
	for _, line := range contentLines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		return strings.HasPrefix(trimmedLine, "{") || strings.HasPrefix(trimmedLine, "[")
	}
	return false
}

func fitANSILineToWidth(value string, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}
	normalizedValue := strings.ReplaceAll(value, "\r", "")
	normalizedValue = strings.ReplaceAll(normalizedValue, "\t", "    ")
	tail := "..."
	if maxWidth <= 3 {
		tail = ""
	}
	truncatedValue := ansi.Truncate(normalizedValue, maxWidth, tail)
	currentWidth := ansi.StringWidth(truncatedValue)
	if currentWidth >= maxWidth {
		return truncatedValue
	}
	return truncatedValue + strings.Repeat(" ", maxWidth-currentWidth)
}

func highlightLogSeverityKeywords(value string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	withWarnings := logWarningPattern.ReplaceAllStringFunc(value, func(matchedValue string) string {
		return "\x1b[33m" + matchedValue + "\x1b[0m"
	})
	return logErrorWordPattern.ReplaceAllStringFunc(withWarnings, func(matchedValue string) string {
		return "\x1b[31m" + matchedValue + "\x1b[0m"
	})
}
