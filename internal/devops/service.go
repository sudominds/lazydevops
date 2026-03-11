package devops

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Pipeline struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Folder    string            `json:"folder"`
	LatestRun PipelineLatestRun `json:"latestRun"`
}

type PipelineLatestRun struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	State        string `json:"state"`
	Result       string `json:"result"`
	SourceBranch string `json:"sourceBranch"`
	CreatedDate  string `json:"createdDate"`
	CreatedOn    string `json:"createdOn"`
	QueueTime    string `json:"queueTime"`
	StartTime    string `json:"startTime"`
	FinishTime   string `json:"finishTime"`
	FinishedDate string `json:"finishedDate"`
}

type Run struct {
	ID            int    `json:"id"`
	BuildNumber   string `json:"buildNumber"`
	Name          string `json:"name"`
	State         string `json:"state"`
	Status        string `json:"status"`
	Result        string `json:"result"`
	Reason        string `json:"reason"`
	SourceVersion string `json:"sourceVersion"`
	SourceBranch  string `json:"sourceBranch"`
	CreatedDate   string `json:"createdDate"`
	CreatedOn     string `json:"createdOn"`
	QueueTime     string `json:"queueTime"`
	StartTime     string `json:"startTime"`
	FinishTime    string `json:"finishTime"`
	FinishedDate  string `json:"finishedDate"`
	WebURL        string `json:"url"`
}

type BuildDetails struct {
	ID            int              `json:"id"`
	BuildNumber   string           `json:"buildNumber"`
	Status        string           `json:"status"`
	Result        string           `json:"result"`
	Reason        string           `json:"reason"`
	QueueTime     string           `json:"queueTime"`
	StartTime     string           `json:"startTime"`
	FinishTime    string           `json:"finishTime"`
	SourceBranch  string           `json:"sourceBranch"`
	SourceVersion string           `json:"sourceVersion"`
	WebURL        string           `json:"url"`
	RequestedFor  Identity         `json:"requestedFor"`
	Definition    PipelineIdentity `json:"definition"`
}

type BuildTimeline struct {
	ID      string           `json:"id"`
	Records []TimelineRecord `json:"records"`
}

type LooseString string

func (value *LooseString) UnmarshalJSON(payload []byte) error {
	var genericValue interface{}
	if unmarshalError := json.Unmarshal(payload, &genericValue); unmarshalError != nil {
		return unmarshalError
	}

	switch typedValue := genericValue.(type) {
	case nil:
		*value = ""
	case string:
		*value = LooseString(strings.TrimSpace(typedValue))
	case float64:
		*value = LooseString(strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.15g", typedValue), "0"), "."))
	case bool:
		*value = LooseString(fmt.Sprintf("%t", typedValue))
	default:
		jsonPayload, marshalError := json.Marshal(typedValue)
		if marshalError != nil {
			return marshalError
		}
		*value = LooseString(strings.TrimSpace(string(jsonPayload)))
	}

	return nil
}

type TimelineRecord struct {
	ID         LooseString          `json:"id"`
	ParentID   LooseString          `json:"parentId"`
	Type       string               `json:"type"`
	Name       string               `json:"name"`
	State      string               `json:"state"`
	Result     string               `json:"result"`
	StartTime  string               `json:"startTime"`
	FinishTime string               `json:"finishTime"`
	Attempt    LooseString          `json:"attempt"`
	Order      LooseString          `json:"order"`
	Log        TimelineLogReference `json:"log"`
	Details    TimelineDetailsRef   `json:"details"`
}

type TimelineLogReference struct {
	ID  LooseString `json:"id"`
	URL string      `json:"url"`
}

type TimelineDetailsRef struct {
	ID  LooseString `json:"id"`
	URL string      `json:"url"`
}

type BuildLog struct {
	ID            LooseString `json:"id"`
	Type          string      `json:"type"`
	URL           string      `json:"url"`
	LineCount     LooseString `json:"lineCount"`
	CreatedOn     string      `json:"createdOn"`
	LastChangedOn string      `json:"lastChangedOn"`
}

type RunDetails struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	State         string           `json:"state"`
	Result        string           `json:"result"`
	Reason        string           `json:"reason"`
	CreatedDate   string           `json:"createdDate"`
	QueueTime     string           `json:"queueTime"`
	StartTime     string           `json:"startTime"`
	FinishedAt    string           `json:"finishedDate"`
	WebURL        string           `json:"url"`
	SourceBranch  string           `json:"sourceBranch"`
	SourceVersion string           `json:"sourceVersion"`
	RequestedBy   Identity         `json:"requestedBy"`
	Pipeline      PipelineIdentity `json:"pipeline"`
	Definition    PipelineIdentity `json:"definition"`
}

type Identity struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

type PipelineIdentity struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Service struct {
	organizationURL string
}

func NewService(organizationURL string) Service {
	return Service{organizationURL: strings.TrimSpace(organizationURL)}
}

func ResolveDefaultOrganization() (string, error) {
	command := exec.Command("az", "devops", "configure", "--list")
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		outputDetails := strings.TrimSpace(string(commandOutput))
		if outputDetails == "" {
			outputDetails = commandError.Error()
		}
		return "", fmt.Errorf("failed to resolve Azure DevOps defaults: %s", outputDetails)
	}

	organizationPattern := regexp.MustCompile(`(?m)^\s*organization\s*=\s*(\S+)\s*$`)
	matches := organizationPattern.FindStringSubmatch(string(commandOutput))
	if len(matches) < 2 {
		return "", fmt.Errorf("no default Azure DevOps organization configured")
	}

	return strings.TrimSpace(matches[1]), nil
}

func (service Service) ListProjects() ([]Project, error) {
	command := exec.Command(
		"az",
		"devops",
		"project",
		"list",
		"--organization",
		service.organizationURL,
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return nil, cliError("list projects", commandError, commandOutput)
	}

	return decodeListPayload[Project](commandOutput)
}

func (service Service) ListPipelines(projectName string) ([]Pipeline, error) {
	command := exec.Command(
		"az",
		"pipelines",
		"list",
		"--organization",
		service.organizationURL,
		"--project",
		projectName,
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return nil, cliError("list pipelines", commandError, commandOutput)
	}

	return decodeListPayload[Pipeline](commandOutput)
}

func (service Service) ListRuns(projectName string, pipelineID int, top int, result string) ([]Run, error) {
	normalizedResult := normalizeRunResultFilter(result)
	pipelineRuns, runsListError := service.listRunsViaRunsList(projectName, pipelineID, top, normalizedResult)
	buildRuns, buildListError := service.listRunsViaBuildList(projectName, pipelineID, top, normalizedResult)

	if buildListError != nil && runsListError != nil {
		return nil, fmt.Errorf("failed to list runs from runs list: %v; failed to list runs from build list: %v", runsListError, buildListError)
	}

	return mergeRunsByID(pipelineRuns, buildRuns), nil
}

func (service Service) listRunsViaBuildList(projectName string, pipelineID int, top int, result string) ([]Run, error) {
	commandArgs := []string{
		"pipelines",
		"build",
		"list",
		"--organization",
		service.organizationURL,
		"--project",
		projectName,
		"--definition-ids",
		fmt.Sprintf("%d", pipelineID),
	}
	normalizedResult := normalizeRunResultFilter(result)
	if normalizedResult != "" {
		commandArgs = append(commandArgs, "--result", normalizedResult)
	}
	commandArgs = append(commandArgs,
		"--query-order",
		"QueueTimeDesc",
		"--top",
		fmt.Sprintf("%d", top),
		"--output",
		"json",
	)
	command := exec.Command("az", commandArgs...)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return nil, cliError("list runs", commandError, commandOutput)
	}

	return decodeListPayload[Run](commandOutput)
}

func (service Service) listRunsViaRunsList(projectName string, pipelineID int, top int, result string) ([]Run, error) {
	commandArgs := []string{
		"pipelines",
		"runs",
		"list",
		"--organization",
		service.organizationURL,
		"--project",
		projectName,
		"--pipeline-ids",
		fmt.Sprintf("%d", pipelineID),
	}
	normalizedResult := normalizeRunResultFilter(result)
	if normalizedResult != "" {
		commandArgs = append(commandArgs, "--result", normalizedResult)
	}
	commandArgs = append(commandArgs,
		"--query-order",
		"QueueTimeDesc",
		"--top",
		fmt.Sprintf("%d", top),
		"--output",
		"json",
	)
	command := exec.Command("az", commandArgs...)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return nil, cliError("list runs", commandError, commandOutput)
	}

	return decodeListPayload[Run](commandOutput)
}

func (service Service) GetRunDetails(projectName string, runID int) (RunDetails, error) {
	command := exec.Command(
		"az",
		"pipelines",
		"runs",
		"show",
		"--organization",
		service.organizationURL,
		"--project",
		projectName,
		"--id",
		fmt.Sprintf("%d", runID),
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return RunDetails{}, cliError("get run details", commandError, commandOutput)
	}

	var runDetails RunDetails
	if parseError := json.Unmarshal(commandOutput, &runDetails); parseError != nil {
		return RunDetails{}, fmt.Errorf("failed to parse run details: %w", parseError)
	}

	return runDetails, nil
}

func (service Service) GetBuildDetails(projectName string, buildID int) (BuildDetails, error) {
	command := exec.Command(
		"az",
		"pipelines",
		"build",
		"show",
		"--organization",
		service.organizationURL,
		"--project",
		projectName,
		"--id",
		fmt.Sprintf("%d", buildID),
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return BuildDetails{}, cliError("get build details", commandError, commandOutput)
	}

	var buildDetails BuildDetails
	if parseError := json.Unmarshal(commandOutput, &buildDetails); parseError != nil {
		return BuildDetails{}, fmt.Errorf("failed to parse build details: %w", parseError)
	}

	return buildDetails, nil
}

func (service Service) GetRunDetailsBuildFirst(projectName string, runID int) (RunDetails, error) {
	buildDetails, buildError := service.GetBuildDetails(projectName, runID)
	runDetails, runError := service.GetRunDetails(projectName, runID)

	if buildError != nil && runError != nil {
		return RunDetails{}, fmt.Errorf("build details error: %v; run details error: %v", buildError, runError)
	}

	mergedDetails := RunDetails{}
	if runError == nil {
		mergedDetails = runDetails
	}
	if buildError == nil {
		mergedDetails = mergeRunDetails(mergedDetails, runDetailsFromBuildDetails(buildDetails))
	}

	return mergedDetails, nil
}

func (service Service) GetBuildTimeline(projectName string, buildID int) (BuildTimeline, error) {
	command := exec.Command(
		"az",
		"devops",
		"invoke",
		"--organization",
		service.organizationURL,
		"--area",
		"build",
		"--resource",
		"timeline",
		"--route-parameters",
		fmt.Sprintf("project=%s", projectName),
		fmt.Sprintf("buildId=%d", buildID),
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return BuildTimeline{}, cliError("get build timeline", commandError, commandOutput)
	}

	var timeline BuildTimeline
	if parseError := json.Unmarshal(commandOutput, &timeline); parseError != nil {
		return BuildTimeline{}, fmt.Errorf("failed to parse build timeline: %w", parseError)
	}

	return timeline, nil
}

func (service Service) GetBuildTimelineDetails(projectName string, buildID int, timelineID string, detailsURL string) (BuildTimeline, error) {
	trimmedTimelineID := strings.TrimSpace(timelineID)
	if trimmedTimelineID != "" {
		childTimeline, childError := service.getBuildTimelineByID(projectName, buildID, trimmedTimelineID)
		if childError == nil {
			return childTimeline, nil
		}
	}
	trimmedDetailsURL := strings.TrimSpace(detailsURL)
	if trimmedDetailsURL != "" {
		childTimeline, childError := service.getBuildTimelineByURL(trimmedDetailsURL)
		if childError == nil {
			return childTimeline, nil
		}
	}
	return BuildTimeline{}, fmt.Errorf("failed to load child timeline for timelineId=%s", trimmedTimelineID)
}

func (service Service) getBuildTimelineByID(projectName string, buildID int, timelineID string) (BuildTimeline, error) {
	command := exec.Command(
		"az",
		"devops",
		"invoke",
		"--organization",
		service.organizationURL,
		"--area",
		"build",
		"--resource",
		"timeline",
		"--route-parameters",
		fmt.Sprintf("project=%s", projectName),
		fmt.Sprintf("buildId=%d", buildID),
		fmt.Sprintf("timelineId=%s", strings.TrimSpace(timelineID)),
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return BuildTimeline{}, cliError("get child build timeline", commandError, commandOutput)
	}

	var timeline BuildTimeline
	if parseError := json.Unmarshal(commandOutput, &timeline); parseError != nil {
		return BuildTimeline{}, fmt.Errorf("failed to parse child build timeline: %w", parseError)
	}
	return timeline, nil
}

func (service Service) getBuildTimelineByURL(detailsURL string) (BuildTimeline, error) {
	command := exec.Command(
		"az",
		"rest",
		"--method",
		"get",
		"--url",
		strings.TrimSpace(detailsURL),
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return BuildTimeline{}, cliError("get child build timeline by url", commandError, commandOutput)
	}

	var timeline BuildTimeline
	if parseError := json.Unmarshal(commandOutput, &timeline); parseError != nil {
		return BuildTimeline{}, fmt.Errorf("failed to parse child build timeline by url: %w", parseError)
	}
	return timeline, nil
}

func (service Service) ListBuildLogs(projectName string, buildID int) ([]BuildLog, error) {
	command := exec.Command(
		"az",
		"devops",
		"invoke",
		"--organization",
		service.organizationURL,
		"--area",
		"build",
		"--resource",
		"logs",
		"--route-parameters",
		fmt.Sprintf("project=%s", projectName),
		fmt.Sprintf("buildId=%d", buildID),
		"--output",
		"json",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return nil, cliError("list build logs", commandError, commandOutput)
	}

	return decodeListPayload[BuildLog](commandOutput)
}

func (service Service) GetBuildLog(projectName string, buildID int, logID int) (string, error) {
	command := exec.Command(
		"az",
		"devops",
		"invoke",
		"--organization",
		service.organizationURL,
		"--area",
		"build",
		"--resource",
		"logs",
		"--route-parameters",
		fmt.Sprintf("project=%s", projectName),
		fmt.Sprintf("buildId=%d", buildID),
		fmt.Sprintf("logId=%d", logID),
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		return "", cliError("get build log", commandError, commandOutput)
	}

	return strings.TrimSpace(string(commandOutput)), nil
}

func runDetailsFromBuildDetails(buildDetails BuildDetails) RunDetails {
	runName := strings.TrimSpace(buildDetails.BuildNumber)
	if runName == "" && buildDetails.ID > 0 {
		runName = fmt.Sprintf("Run %d", buildDetails.ID)
	}

	createdDate := firstNonEmptyString(buildDetails.QueueTime, buildDetails.StartTime, buildDetails.FinishTime)
	return RunDetails{
		ID:            buildDetails.ID,
		Name:          runName,
		State:         buildDetails.Status,
		Result:        buildDetails.Result,
		Reason:        buildDetails.Reason,
		CreatedDate:   createdDate,
		QueueTime:     buildDetails.QueueTime,
		StartTime:     buildDetails.StartTime,
		FinishedAt:    buildDetails.FinishTime,
		WebURL:        buildDetails.WebURL,
		SourceBranch:  buildDetails.SourceBranch,
		SourceVersion: buildDetails.SourceVersion,
		RequestedBy:   buildDetails.RequestedFor,
		Pipeline:      buildDetails.Definition,
		Definition:    buildDetails.Definition,
	}
}

func mergeRunDetails(base RunDetails, overlay RunDetails) RunDetails {
	merged := base
	merged.ID = firstNonZeroInt(overlay.ID, base.ID)
	merged.Name = firstNonEmptyString(overlay.Name, base.Name)
	merged.State = firstNonEmptyString(overlay.State, base.State)
	merged.Result = firstNonEmptyString(overlay.Result, base.Result)
	merged.Reason = firstNonEmptyString(overlay.Reason, base.Reason)
	merged.CreatedDate = firstNonEmptyString(overlay.CreatedDate, base.CreatedDate)
	merged.QueueTime = firstNonEmptyString(overlay.QueueTime, base.QueueTime)
	merged.StartTime = firstNonEmptyString(overlay.StartTime, base.StartTime)
	merged.FinishedAt = firstNonEmptyString(overlay.FinishedAt, base.FinishedAt)
	merged.WebURL = firstNonEmptyString(overlay.WebURL, base.WebURL)
	merged.SourceBranch = firstNonEmptyString(overlay.SourceBranch, base.SourceBranch)
	merged.SourceVersion = firstNonEmptyString(overlay.SourceVersion, base.SourceVersion)
	merged.RequestedBy = chooseIdentity(overlay.RequestedBy, base.RequestedBy)
	merged.Pipeline = choosePipelineIdentity(overlay.Pipeline, base.Pipeline)
	merged.Definition = choosePipelineIdentity(overlay.Definition, base.Definition)
	return merged
}

func chooseIdentity(preferred Identity, fallback Identity) Identity {
	if strings.TrimSpace(preferred.DisplayName) != "" || strings.TrimSpace(preferred.UniqueName) != "" {
		return preferred
	}
	return fallback
}

func choosePipelineIdentity(preferred PipelineIdentity, fallback PipelineIdentity) PipelineIdentity {
	if preferred.ID > 0 || strings.TrimSpace(preferred.Name) != "" {
		return preferred
	}
	return fallback
}

func firstNonZeroInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue != "" {
			return trimmedValue
		}
	}
	return ""
}

func normalizeRunResultFilter(value string) string {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	normalizedValue = strings.NewReplacer(" ", "", "-", "", "_", "").Replace(normalizedValue)
	switch normalizedValue {
	case "canceled", "cancelled", "failed", "none", "partiallysucceeded", "succeeded":
		return map[string]string{
			"canceled":           "canceled",
			"cancelled":          "canceled",
			"failed":             "failed",
			"none":               "none",
			"partiallysucceeded": "partiallySucceeded",
			"succeeded":          "succeeded",
		}[normalizedValue]
	default:
		return ""
	}
}

func mergeRunsByID(primaryRuns []Run, secondaryRuns []Run) []Run {
	mergedByID := map[int]Run{}
	orderedIDs := make([]int, 0, len(primaryRuns)+len(secondaryRuns))

	upsert := func(runEntry Run) {
		if runEntry.ID <= 0 {
			return
		}
		existingRun, exists := mergedByID[runEntry.ID]
		if !exists {
			mergedByID[runEntry.ID] = runEntry
			orderedIDs = append(orderedIDs, runEntry.ID)
			return
		}
		mergedByID[runEntry.ID] = mergeRunEntry(existingRun, runEntry)
	}

	for _, runEntry := range primaryRuns {
		upsert(runEntry)
	}
	for _, runEntry := range secondaryRuns {
		upsert(runEntry)
	}

	mergedRuns := make([]Run, 0, len(orderedIDs))
	for _, runID := range orderedIDs {
		if runEntry, exists := mergedByID[runID]; exists {
			mergedRuns = append(mergedRuns, runEntry)
		}
	}

	return mergedRuns
}

func mergeRunEntry(base Run, overlay Run) Run {
	merged := base
	merged.BuildNumber = firstNonEmptyString(base.BuildNumber, overlay.BuildNumber)
	merged.Name = firstNonEmptyString(base.Name, overlay.Name)
	merged.State = firstNonEmptyString(base.State, overlay.State)
	merged.Status = firstNonEmptyString(base.Status, overlay.Status)
	merged.Result = firstNonEmptyString(base.Result, overlay.Result)
	merged.Reason = firstNonEmptyString(base.Reason, overlay.Reason)
	merged.SourceVersion = firstNonEmptyString(base.SourceVersion, overlay.SourceVersion)
	merged.SourceBranch = firstNonEmptyString(base.SourceBranch, overlay.SourceBranch)
	merged.CreatedDate = firstNonEmptyString(base.CreatedDate, overlay.CreatedDate)
	merged.CreatedOn = firstNonEmptyString(base.CreatedOn, overlay.CreatedOn)
	merged.QueueTime = firstNonEmptyString(base.QueueTime, overlay.QueueTime)
	merged.StartTime = firstNonEmptyString(base.StartTime, overlay.StartTime)
	merged.FinishTime = firstNonEmptyString(base.FinishTime, overlay.FinishTime)
	merged.FinishedDate = firstNonEmptyString(base.FinishedDate, overlay.FinishedDate)
	merged.WebURL = firstNonEmptyString(base.WebURL, overlay.WebURL)
	return merged
}

func decodeListPayload[entityType any](payload []byte) ([]entityType, error) {
	var directList []entityType
	if parseError := json.Unmarshal(payload, &directList); parseError == nil {
		return directList, nil
	}

	var envelope struct {
		Value []entityType `json:"value"`
	}
	if parseError := json.Unmarshal(payload, &envelope); parseError != nil {
		return nil, fmt.Errorf("failed to parse list payload: %w", parseError)
	}

	return envelope.Value, nil
}

func cliError(action string, commandError error, commandOutput []byte) error {
	outputDetails := strings.TrimSpace(string(commandOutput))
	if outputDetails == "" {
		outputDetails = commandError.Error()
	}
	return fmt.Errorf("failed to %s: %s", action, outputDetails)
}
