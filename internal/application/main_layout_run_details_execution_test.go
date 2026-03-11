package application

import (
	"testing"

	"lazydevops/internal/devops"
)

func TestParseLooseOrderSupportsDecimalValues(t *testing.T) {
	order, hasOrder := parseLooseOrder("2.5")
	if !hasOrder {
		t.Fatalf("expected decimal order to parse")
	}
	if order != 2.5 {
		t.Fatalf("unexpected parsed order: got %v want 2.5", order)
	}
}

func TestSummarizeExecutionSortsJobsByNumericOrder(t *testing.T) {
	timeline := devops.BuildTimeline{
		Records: []devops.TimelineRecord{
			{
				ID:     devops.LooseString("job-a"),
				Type:   "job",
				Name:   "Task Group Beta",
				Order:  devops.LooseString("2.5"),
				Result: "succeeded",
				State:  "completed",
			},
			{
				ID:     devops.LooseString("job-b"),
				Type:   "job",
				Name:   "Task Group Alpha",
				Order:  devops.LooseString("2"),
				Result: "succeeded",
				State:  "completed",
			},
		},
	}

	_, jobs, _ := summarizeExecution(timeline)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].Name != "Task Group Alpha" {
		t.Fatalf("expected Task Group Alpha first, got %q", jobs[0].Name)
	}
	if jobs[1].Name != "Task Group Beta" {
		t.Fatalf("expected Task Group Beta second, got %q", jobs[1].Name)
	}
}
