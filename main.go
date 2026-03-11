package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"lazydevops/internal/application"
	"lazydevops/internal/devops"
	"lazydevops/internal/onboarding"
)

func main() {
	initialModel := resolveInitialModel()
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func resolveInitialModel() tea.Model {
	if onboarding.IsCompleted() {
		organizationURL, resolveError := devops.ResolveDefaultOrganization()
		if resolveError == nil {
			return application.NewMainLayoutModel(organizationURL)
		}
	}

	return application.NewSetupModel()
}
