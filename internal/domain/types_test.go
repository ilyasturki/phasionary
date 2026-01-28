package domain

import (
	"reflect"
	"testing"
)

func TestSortTasks(t *testing.T) {
	tasks := []Task{
		{Title: "Alpha", Priority: PriorityHigh},
		{Title: "Beta", Priority: PriorityHigh},
		{Title: "Gamma", Priority: PriorityHigh},
		{Title: "Delta", Priority: PriorityMedium},
		{Title: "Echo", Priority: PriorityMedium},
		{Title: "Foxtrot", Priority: PriorityMedium},
		{Title: "aardvark", Priority: PriorityMedium},
	}

	SortTasks(tasks)

	got := make([]string, 0, len(tasks))
	for _, task := range tasks {
		got = append(got, task.Title)
	}

	want := []string{
		"Alpha",
		"Beta",
		"Gamma",
		"aardvark",
		"Delta",
		"Echo",
		"Foxtrot",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sort order: got %v, want %v", got, want)
	}
}
