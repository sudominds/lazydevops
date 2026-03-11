package application

import (
	"fmt"
	"strings"

	"lazydevops/internal/application/components/search"
	"lazydevops/internal/devops"
)

func (model MainLayoutModel) filteredItemIndexes() []int {
	searchTerm := model.searchInput.Value()

	matches := make([]int, 0)
	switch model.currentStage {
	case stageProjects:
		for index, project := range model.projects {
			if search.Matches(project.Name, searchTerm) {
				matches = append(matches, index)
			}
		}
	case stagePipelines:
		for index, pipeline := range model.pipelines {
			pipelineLabel := fmt.Sprintf("%s %d", pipeline.Name, pipeline.ID)
			if search.Matches(pipelineLabel, searchTerm) {
				matches = append(matches, index)
			}
		}
	case stageRuns:
		runFilter := parseRunSearchFilter(searchTerm)
		for index, run := range model.runs {
			if !runMatchesResultFilter(run, runFilter.ResultValues) {
				continue
			}
			runLabel := runSearchLabel(run)
			if search.Matches(runLabel, runFilter.TextQuery) {
				matches = append(matches, index)
			}
		}
	}

	return matches
}

type runSearchFilter struct {
	ResultValues []string
	TextQuery    string
}

func normalizedRunResultFilter(rawValue string) string {
	normalizedValue := strings.ToLower(strings.TrimSpace(rawValue))
	normalizedValue = strings.NewReplacer(" ", "", "-", "", "_", "").Replace(normalizedValue)
	switch normalizedValue {
	case "canceled", "cancelled":
		return "canceled"
	case "failed":
		return "failed"
	case "none":
		return "none"
	case "partiallysucceeded":
		return "partiallySucceeded"
	case "succeeded":
		return "succeeded"
	default:
		return ""
	}
}

func (model MainLayoutModel) activeRunResultFilterLabel() string {
	normalizedValue := normalizedRunResultFilter(model.runsResultFilter)
	if normalizedValue == "" {
		return "any"
	}
	return normalizedValue
}

func parseRunSearchFilter(rawQuery string) runSearchFilter {
	parts := strings.Fields(strings.TrimSpace(rawQuery))
	resultValues := make([]string, 0, len(parts))
	textParts := make([]string, 0, len(parts))
	for _, part := range parts {
		loweredPart := strings.ToLower(strings.TrimSpace(part))
		if !strings.HasPrefix(loweredPart, "result:") && !strings.HasPrefix(loweredPart, "status:") {
			textParts = append(textParts, part)
			continue
		}
		tokenParts := strings.SplitN(part, ":", 2)
		if len(tokenParts) != 2 {
			continue
		}
		resultValue := normalizedRunResultFilter(tokenParts[1])
		if resultValue == "" {
			continue
		}
		resultValues = append(resultValues, resultValue)
	}
	return runSearchFilter{
		ResultValues: resultValues,
		TextQuery:    strings.Join(textParts, " "),
	}
}

func runSearchLabel(run devops.Run) string {
	return strings.Join([]string{
		runDisplayName(run),
		run.BuildNumber,
		fmt.Sprintf("%d", run.ID),
		run.State,
		run.Result,
		run.Status,
		run.Reason,
		formatRunBranch(run.SourceBranch),
		shortCommitHash(run.SourceVersion),
	}, " ")
}

func runMatchesResultFilter(run devops.Run, expectedResults []string) bool {
	if len(expectedResults) == 0 {
		return true
	}
	resultValue := normalizeStatusKey(run.Result)
	for _, expectedResult := range expectedResults {
		if normalizeStatusKey(expectedResult) == resultValue {
			return true
		}
	}
	return false
}

func (model MainLayoutModel) breadcrumb() string {
	path := []string{"Projects"}

	selectedProject, hasProject := model.selectedProject()
	if hasProject && (model.currentStage == stagePipelines || model.currentStage == stageRuns || model.currentStage == stageRunDetails) {
		path = append(path, selectedProject.Name)
	}

	selectedPipeline, hasPipeline := model.selectedPipeline()
	if hasPipeline && (model.currentStage == stageRuns || model.currentStage == stageRunDetails) {
		path = append(path, selectedPipeline.Name)
	}

	selectedRun, hasRun := model.selectedRun()
	if hasRun && model.currentStage == stageRunDetails {
		path = append(path, runDisplayName(selectedRun))
	}

	return "Path: " + strings.Join(path, " -> ")
}

func (model MainLayoutModel) helpText() string {
	return "esc: back/close (run details back: b) | runs result picker: alt+f | keymap: ? | run details: tab,left,right,h,l,0,2,ctrl+d,ctrl+u,gg,G | command: : | quit: :q"
}

func (model MainLayoutModel) currentAZCommandPreview() string {
	switch model.currentStage {
	case stageProjects:
		return fmt.Sprintf("az devops project list --organization %s --output json", shellSingleQuote(model.organizationURL))
	case stagePipelines:
		selectedProject, hasProject := model.selectedProject()
		projectName := "<project>"
		if hasProject {
			projectName = selectedProject.Name
		}
		return buildPipelinesListCommand(model.organizationURL, projectName)
	case stageRuns:
		selectedProject, hasProject := model.selectedProject()
		selectedPipeline, hasPipeline := model.selectedPipeline()
		projectName := "<project>"
		pipelineID := "<pipeline-id>"
		if hasProject {
			projectName = selectedProject.Name
		}
		if hasPipeline {
			pipelineID = fmt.Sprintf("%d", selectedPipeline.ID)
		}
		return buildRunsListCommand(model.organizationURL, projectName, pipelineID, normalizedRunResultFilter(model.runsResultFilter))
	case stageRunDetails:
		projectName, runID := model.runDetailsCommandContext()
		currentLogs := model.currentRunDetailsLogs()
		if len(currentLogs) > 0 {
			selectedLog := currentLogs[clampSelection(model.runDetailsLogsCursor, len(currentLogs))]
			logID, hasNumericLogID := parseLooseInt(string(selectedLog.ID))
			if hasNumericLogID {
				return buildRunLogCommand(model.organizationURL, projectName, runID, logID)
			}
		}
		if model.hasRunDetails {
			return buildRunTimelineCommand(model.organizationURL, projectName, runID)
		}
		return buildRunDetailsCommand(model.organizationURL, projectName, runID)
	default:
		return fmt.Sprintf("az devops project list --organization %s --output json", shellSingleQuote(model.organizationURL))
	}
}
