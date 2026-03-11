package onboarding

import (
	"os"
	"path/filepath"
)

const (
	applicationDirectoryName = ".lazydevops"
	completedMarkerFileName  = "onboarding_complete"
)

func IsCompleted() bool {
	markerPath, resolveError := markerFilePath()
	if resolveError != nil {
		return false
	}

	_, statError := os.Stat(markerPath)
	return statError == nil
}

func MarkCompleted() error {
	markerPath, resolveError := markerFilePath()
	if resolveError != nil {
		return resolveError
	}

	if makeDirectoryError := os.MkdirAll(filepath.Dir(markerPath), 0o755); makeDirectoryError != nil {
		return makeDirectoryError
	}

	return os.WriteFile(markerPath, []byte("completed\n"), 0o644)
}

func markerFilePath() (string, error) {
	homeDirectory, homeError := os.UserHomeDir()
	if homeError != nil {
		return "", homeError
	}

	return filepath.Join(homeDirectory, applicationDirectoryName, completedMarkerFileName), nil
}
