package application

import (
	"testing"

	"lazydevops/internal/devops"
)

func TestBuildRunDetailsStageGroupsUsesTopLevelStageJobsInOrder(t *testing.T) {
	model := NewMainLayoutModel("https://dev.azure.com/example")
	model.selectedRunTimeline = devops.BuildTimeline{
		Records: []devops.TimelineRecord{
			{
				ID:    devops.LooseString("stage-1"),
				Type:  "stage",
				Name:  "Stage Alpha",
				Order: devops.LooseString("1"),
			},
			{
				ID:       devops.LooseString("job-build"),
				ParentID: devops.LooseString("stage-1"),
				Type:     "job",
				Name:     "Job Alpha",
				Order:    devops.LooseString("1"),
			},
			{
				ID:       devops.LooseString("job-build-child"),
				ParentID: devops.LooseString("job-build"),
				Type:     "job",
				Name:     "Job Alpha Child",
				Order:    devops.LooseString("1.5"),
			},
			{
				ID:       devops.LooseString("job-nservicebus"),
				ParentID: devops.LooseString("stage-1"),
				Type:     "job",
				Name:     "Job Beta",
				Order:    devops.LooseString("2"),
			},
			{
				ID:       devops.LooseString("job-ef"),
				ParentID: devops.LooseString("stage-1"),
				Type:     "job",
				Name:     "Job Gamma",
				Order:    devops.LooseString("3"),
			},
			{
				ID:       devops.LooseString("job-tests"),
				ParentID: devops.LooseString("stage-1"),
				Type:     "job",
				Name:     "Job Delta",
				Order:    devops.LooseString("4"),
			},
			{
				ID:       devops.LooseString("job-publish"),
				ParentID: devops.LooseString("stage-1"),
				Type:     "job",
				Name:     "Job Epsilon",
				Order:    devops.LooseString("5"),
			},
		},
	}

	groups := model.buildRunDetailsStageGroups()
	if len(groups) != 1 {
		t.Fatalf("expected one stage group, got %d", len(groups))
	}
	if len(groups[0].Jobs) != 5 {
		t.Fatalf("expected 5 top-level jobs, got %d", len(groups[0].Jobs))
	}

	expectedOrder := []string{
		"Job Alpha",
		"Job Beta",
		"Job Gamma",
		"Job Delta",
		"Job Epsilon",
	}
	for index, expectedName := range expectedOrder {
		if groups[0].Jobs[index].Job.Name != expectedName {
			t.Fatalf("unexpected job at index %d: got %q want %q", index, groups[0].Jobs[index].Job.Name, expectedName)
		}
	}
}

func TestTaskSummaryLessUsesStartTimeBeforeOrder(t *testing.T) {
	earlierTask := taskSummary{
		Name:     "Task Alpha",
		StartAt:  "2026-03-09T10:00:00Z",
		Order:    20,
		HasOrder: true,
		Sequence: 2,
	}
	laterTask := taskSummary{
		Name:     "Task Beta",
		StartAt:  "2026-03-09T10:02:00Z",
		Order:    1,
		HasOrder: true,
		Sequence: 1,
	}

	if !taskSummaryLess(earlierTask, laterTask) {
		t.Fatalf("expected earlier task start time to sort first")
	}
	if taskSummaryLess(laterTask, earlierTask) {
		t.Fatalf("expected later task start time to sort after earlier task")
	}
}
