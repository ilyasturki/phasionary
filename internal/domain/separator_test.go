package domain

import "testing"

func TestNewSeparatorIsSeparator(t *testing.T) {
	sep, err := NewSeparator()
	if err != nil {
		t.Fatalf("NewSeparator: %v", err)
	}
	if !sep.IsSeparator() {
		t.Fatalf("expected IsSeparator() true, got kind %q", sep.Kind)
	}
	if sep.ID == "" {
		t.Fatal("expected separator to have an ID")
	}
	ordinary, _ := NewTask("real")
	if ordinary.IsSeparator() {
		t.Fatal("ordinary task reported as separator")
	}
}

func TestSeparatorsExcludedFromTallies(t *testing.T) {
	cat := Category{
		Tasks: []Task{
			{Title: "one", Status: StatusTodo},
			{Kind: KindSeparator, Title: "divider"},
			{Title: "two", Status: StatusCompleted},
		},
	}

	counts := cat.StatusCounts()
	if counts.Total() != 2 {
		t.Fatalf("StatusCounts.Total() = %d, want 2 (separator must not count)", counts.Total())
	}
	if counts.Todo != 1 || counts.Completed != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
	if got := cat.TaskCount(); got != 2 {
		t.Fatalf("TaskCount() = %d, want 2", got)
	}
}

func TestAggregateStatusIgnoresSeparators(t *testing.T) {
	// A category holding only separators has no aggregate status.
	only := Category{Tasks: []Task{{Kind: KindSeparator}, {Kind: KindSeparator, Title: "x"}}}
	if got := only.AggregateStatus(); got != "" {
		t.Fatalf("AggregateStatus() = %q, want empty for separator-only category", got)
	}

	// Separators don't drag an all-completed category out of the completed state.
	done := Category{Tasks: []Task{
		{Status: StatusCompleted},
		{Kind: KindSeparator},
		{Status: StatusCompleted},
	}}
	if got := done.AggregateStatus(); got != StatusCompleted {
		t.Fatalf("AggregateStatus() = %q, want %q", got, StatusCompleted)
	}
}
