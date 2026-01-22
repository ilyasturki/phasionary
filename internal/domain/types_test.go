package domain

import (
	"reflect"
	"testing"
)

func TestSortTasks(t *testing.T) {
	tasks := []Task{
		{Title: "Alpha", Priority: PriorityHigh, Deadline: "2026-01-10", TimeEstimateValue: 2, TimeEstimateUnit: EstimateHours},
		{Title: "Beta", Priority: PriorityHigh, Deadline: "2026-01-05", TimeEstimateValue: 4, TimeEstimateUnit: EstimateHours},
		{Title: "Gamma", Priority: PriorityHigh, Deadline: "2026-01-05", TimeEstimateValue: 60, TimeEstimateUnit: EstimateMinutes},
		{Title: "Delta", Priority: PriorityMedium, Deadline: "2026-01-01"},
		{Title: "Echo", Priority: PriorityMedium, TimeEstimateValue: 30, TimeEstimateUnit: EstimateMinutes},
		{Title: "Foxtrot", Priority: PriorityMedium, TimeEstimateValue: 1, TimeEstimateUnit: EstimateHours},
		{Title: "aardvark", Priority: PriorityMedium, TimeEstimateValue: 1, TimeEstimateUnit: EstimateHours},
	}

	SortTasks(tasks)

	got := make([]string, 0, len(tasks))
	for _, task := range tasks {
		got = append(got, task.Title)
	}

	want := []string{
		"Gamma",
		"Beta",
		"Alpha",
		"Delta",
		"Echo",
		"aardvark",
		"Foxtrot",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sort order: got %v, want %v", got, want)
	}
}
