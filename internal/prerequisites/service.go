package prerequisites

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
)

type CheckIdentifier string

const (
	CheckAzureCliInstalled             CheckIdentifier = "azure_cli_installed"
	CheckAzureLoginValid               CheckIdentifier = "azure_login_valid"
	CheckAzureDevOpsExtensionInstalled CheckIdentifier = "azure_devops_extension_installed"
	CheckAzureDevOpsDefaultsConfigured CheckIdentifier = "azure_devops_defaults_configured"
	CheckAzureDevOpsOrganizationAccess CheckIdentifier = "azure_devops_organization_access"
)

type Definition struct {
	Identifier  CheckIdentifier
	Title       string
	Description string
}

type Result struct {
	Identifier  CheckIdentifier
	Title       string
	Description string
	Status      Status
	Details     string
	Remediation []string
}

type Service struct {
	definitions []Definition
}

func NewService() Service {
	return Service{
		definitions: []Definition{
			{
				Identifier:  CheckAzureCliInstalled,
				Title:       "Azure CLI installed",
				Description: "Checks whether 'az' is available on PATH.",
			},
			{
				Identifier:  CheckAzureLoginValid,
				Title:       "Azure login valid",
				Description: "Checks whether the Azure CLI has an active authenticated account.",
			},
			{
				Identifier:  CheckAzureDevOpsExtensionInstalled,
				Title:       "Azure DevOps extension installed",
				Description: "Checks whether the Azure CLI azure-devops extension is installed.",
			},
			{
				Identifier:  CheckAzureDevOpsDefaultsConfigured,
				Title:       "Azure DevOps defaults configured",
				Description: "Checks whether a default Azure DevOps organization is configured.",
			},
			{
				Identifier:  CheckAzureDevOpsOrganizationAccess,
				Title:       "Azure DevOps org access",
				Description: "Checks whether the configured organization can be queried.",
			},
		},
	}
}

func (service Service) Definitions() []Definition {
	copiedDefinitions := make([]Definition, len(service.definitions))
	copy(copiedDefinitions, service.definitions)
	return copiedDefinitions
}

func (service Service) Run(
	checkIdentifier CheckIdentifier,
	previousResults map[CheckIdentifier]Result,
) Result {
	definition := service.definitionByIdentifier(checkIdentifier)

	switch checkIdentifier {
	case CheckAzureCliInstalled:
		return service.runAzureCliInstalledCheck(definition)
	case CheckAzureLoginValid:
		return service.runAzureLoginCheck(definition, previousResults)
	case CheckAzureDevOpsExtensionInstalled:
		return service.runAzureDevOpsExtensionCheck(definition, previousResults)
	case CheckAzureDevOpsDefaultsConfigured:
		return service.runAzureDevOpsDefaultsCheck(definition, previousResults)
	case CheckAzureDevOpsOrganizationAccess:
		return service.runAzureDevOpsOrganizationAccessCheck(definition, previousResults)
	default:
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     fmt.Sprintf("Unknown check identifier: %s", checkIdentifier),
		}
	}
}

func (service Service) ConfigureDefaultOrganization(organizationURL string) error {
	command := exec.Command(
		"az",
		"devops",
		"configure",
		"--defaults",
		fmt.Sprintf("organization=%s", organizationURL),
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		outputDetails := strings.TrimSpace(string(commandOutput))
		if outputDetails == "" {
			outputDetails = commandError.Error()
		}

		return fmt.Errorf("failed to set default organization: %s", outputDetails)
	}

	return nil
}

func (service Service) runAzureCliInstalledCheck(definition Definition) Result {
	azureCliPath, lookupError := exec.LookPath("az")
	if lookupError != nil {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Azure CLI is not installed or not available on PATH.",
			Remediation: []string{
				"Install Azure CLI: https://learn.microsoft.com/cli/azure/install-azure-cli",
				"After install, restart terminal if PATH was updated.",
				"Press 'r' to recheck.",
			},
		}
	}

	return Result{
		Identifier:  definition.Identifier,
		Title:       definition.Title,
		Description: definition.Description,
		Status:      StatusPassed,
		Details:     fmt.Sprintf("Found Azure CLI at: %s", azureCliPath),
	}
}

func (service Service) runAzureLoginCheck(
	definition Definition,
	previousResults map[CheckIdentifier]Result,
) Result {
	azureCliCheckResult, hasAzureCliResult := previousResults[CheckAzureCliInstalled]
	if !hasAzureCliResult || azureCliCheckResult.Status != StatusPassed {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Cannot validate login because Azure CLI is not ready.",
			Remediation: []string{
				"Resolve the Azure CLI installation issue first.",
				"Press 'r' to run checks again.",
			},
		}
	}

	command := exec.Command("az", "account", "show", "--output", "none")
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		outputDetails := strings.TrimSpace(string(commandOutput))
		if outputDetails == "" {
			outputDetails = commandError.Error()
		}

		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Azure CLI is installed but no active login was found.",
			Remediation: []string{
				"Run: az login",
				fmt.Sprintf("CLI output: %s", outputDetails),
				"Press 'r' to recheck.",
			},
		}
	}

	return Result{
		Identifier:  definition.Identifier,
		Title:       definition.Title,
		Description: definition.Description,
		Status:      StatusPassed,
		Details:     "Active Azure CLI login detected.",
	}
}

func (service Service) runAzureDevOpsExtensionCheck(
	definition Definition,
	previousResults map[CheckIdentifier]Result,
) Result {
	if !service.isDependencyPassed(previousResults, CheckAzureCliInstalled) {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Cannot validate extension because Azure CLI is not ready.",
			Remediation: []string{
				"Resolve the Azure CLI installation issue first.",
				"Press 'r' to run checks again.",
			},
		}
	}

	command := exec.Command("az", "extension", "show", "--name", "azure-devops", "--output", "none")
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		outputDetails := strings.TrimSpace(string(commandOutput))
		if outputDetails == "" {
			outputDetails = commandError.Error()
		}

		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Azure DevOps extension is not installed.",
			Remediation: []string{
				"Run: az extension add --name azure-devops",
				fmt.Sprintf("CLI output: %s", outputDetails),
				"Press 'r' to recheck.",
			},
		}
	}

	return Result{
		Identifier:  definition.Identifier,
		Title:       definition.Title,
		Description: definition.Description,
		Status:      StatusPassed,
		Details:     "Azure DevOps extension is installed.",
	}
}

func (service Service) runAzureDevOpsDefaultsCheck(
	definition Definition,
	previousResults map[CheckIdentifier]Result,
) Result {
	if !service.isDependencyPassed(previousResults, CheckAzureCliInstalled) ||
		!service.isDependencyPassed(previousResults, CheckAzureDevOpsExtensionInstalled) {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Cannot validate defaults because Azure CLI prerequisites are not ready.",
			Remediation: []string{
				"Resolve Azure CLI and extension issues first.",
				"Press 'r' to run checks again.",
			},
		}
	}

	command := exec.Command("az", "devops", "configure", "--list")
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		outputDetails := strings.TrimSpace(string(commandOutput))
		if outputDetails == "" {
			outputDetails = commandError.Error()
		}

		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Unable to read Azure DevOps defaults from Azure CLI.",
			Remediation: []string{
				fmt.Sprintf("CLI output: %s", outputDetails),
				"Run: az devops configure --defaults organization=https://dev.azure.com/<your-org>",
				"Press 'r' to recheck.",
			},
		}
	}

	organizationUrl := parseAzureDevOpsOrganizationFromConfig(string(commandOutput))
	if organizationUrl == "" {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "No default Azure DevOps organization is configured.",
			Remediation: []string{
				"Run: az devops configure --defaults organization=https://dev.azure.com/<your-org>",
				"You can also provide --organization explicitly in future commands.",
				"Press 'r' to recheck.",
			},
		}
	}

	return Result{
		Identifier:  definition.Identifier,
		Title:       definition.Title,
		Description: definition.Description,
		Status:      StatusPassed,
		Details:     fmt.Sprintf("Default organization: %s", organizationUrl),
	}
}

func (service Service) runAzureDevOpsOrganizationAccessCheck(
	definition Definition,
	previousResults map[CheckIdentifier]Result,
) Result {
	if !service.isDependencyPassed(previousResults, CheckAzureCliInstalled) ||
		!service.isDependencyPassed(previousResults, CheckAzureLoginValid) ||
		!service.isDependencyPassed(previousResults, CheckAzureDevOpsExtensionInstalled) {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Cannot validate organization access because prerequisites are not ready.",
			Remediation: []string{
				"Resolve Azure CLI, login, and extension checks first.",
				"Press 'r' to run checks again.",
			},
		}
	}

	defaultsResult, hasDefaultsResult := previousResults[CheckAzureDevOpsDefaultsConfigured]
	if !hasDefaultsResult || defaultsResult.Status != StatusPassed {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Cannot validate organization access because no default organization was found.",
			Remediation: []string{
				"Set organization default first: az devops configure --defaults organization=https://dev.azure.com/<your-org>",
				"Press 'r' to recheck.",
			},
		}
	}

	organizationUrl := parseOrganizationFromDetails(defaultsResult.Details)
	if organizationUrl == "" {
		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     "Could not parse organization URL from defaults check output.",
			Remediation: []string{
				"Run: az devops configure --list",
				"Ensure organization is set to a valid https://dev.azure.com/<your-org> URL.",
				"Press 'r' to recheck.",
			},
		}
	}

	command := exec.Command(
		"az",
		"devops",
		"project",
		"list",
		"--organization",
		organizationUrl,
		"--top",
		"1",
		"--output",
		"none",
	)
	commandOutput, commandError := command.CombinedOutput()
	if commandError != nil {
		outputDetails := strings.TrimSpace(string(commandOutput))
		if outputDetails == "" {
			outputDetails = commandError.Error()
		}

		return Result{
			Identifier:  definition.Identifier,
			Title:       definition.Title,
			Description: definition.Description,
			Status:      StatusFailed,
			Details:     fmt.Sprintf("Failed to query organization %s.", organizationUrl),
			Remediation: []string{
				"Check your Azure DevOps permissions for this organization.",
				"Confirm the organization URL is correct.",
				fmt.Sprintf("CLI output: %s", outputDetails),
				"Press 'r' to recheck.",
			},
		}
	}

	return Result{
		Identifier:  definition.Identifier,
		Title:       definition.Title,
		Description: definition.Description,
		Status:      StatusPassed,
		Details:     fmt.Sprintf("Organization access verified for %s.", organizationUrl),
	}
}

func (service Service) definitionByIdentifier(checkIdentifier CheckIdentifier) Definition {
	for _, currentDefinition := range service.definitions {
		if currentDefinition.Identifier == checkIdentifier {
			return currentDefinition
		}
	}

	return Definition{
		Identifier:  checkIdentifier,
		Title:       string(checkIdentifier),
		Description: "No description available.",
	}
}

func (service Service) isDependencyPassed(
	previousResults map[CheckIdentifier]Result,
	dependencyIdentifier CheckIdentifier,
) bool {
	dependencyResult, hasDependencyResult := previousResults[dependencyIdentifier]
	return hasDependencyResult && dependencyResult.Status == StatusPassed
}

func parseAzureDevOpsOrganizationFromConfig(configurationOutput string) string {
	// Typical output line: "organization = https://dev.azure.com/example"
	organizationPattern := regexp.MustCompile(`(?m)^\s*organization\s*=\s*(\S+)\s*$`)
	matches := organizationPattern.FindStringSubmatch(configurationOutput)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(matches[1])
}

func parseOrganizationFromDetails(details string) string {
	detailPrefix := "Default organization: "
	if strings.HasPrefix(details, detailPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(details, detailPrefix))
	}

	return ""
}
