package journal

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
)

func mustTask(t *testing.T, title string) domain.Task {
	t.Helper()
	task, err := domain.NewTask(title)
	require.NoError(t, err)
	return task
}

func testProject(t *testing.T) domain.Project {
	t.Helper()
	p, err := domain.NewProject("Test")
	require.NoError(t, err)
	cat, err := domain.NewCategory("Inbox")
	require.NoError(t, err)
	cat.Tasks = []domain.Task{mustTask(t, "one"), mustTask(t, "two")}
	p.Categories = []domain.Category{cat}
	return p
}

// forkTasks copies p with its own category and first-category task slices, so
// a test can rewrite those tasks without touching p.
func forkTasks(p domain.Project) (domain.Project, []domain.Task) {
	p.Categories = slices.Clone(p.Categories)
	p.Categories[0].Tasks = slices.Clone(p.Categories[0].Tasks)
	return p, p.Categories[0].Tasks
}

func kinds(drafts []Draft) []string {
	out := make([]string, len(drafts))
	for i, d := range drafts {
		out[i] = d.Kind
	}
	return out
}

func TestDiffCreateCascade(t *testing.T) {
	p := testProject(t)
	drafts := DiffProjects(nil, p)
	assert.Equal(t, []string{
		KindProjectCreate,
		KindCategoryCreate,
		KindTaskCreate,
		KindTaskCreate,
		KindCategoryReorder,
		KindProjectReorder,
	}, kinds(drafts))

	assert.Equal(t, p.Name, drafts[0].Fields["name"])
	assert.Equal(t, p.Categories[0].ID, drafts[1].EntityID)
	assert.Equal(t, p.Categories[0].ID, drafts[2].Fields["category_id"])
	assert.Equal(t, "one", drafts[2].Fields["title"])
	assert.Equal(t, domain.StatusTodo, drafts[2].Fields["status"])
	assert.Equal(t,
		[]string{p.Categories[0].Tasks[0].ID, p.Categories[0].Tasks[1].ID},
		drafts[4].Fields["task_ids"])
}

func TestDiffProjectTimestampOnlyIsNoise(t *testing.T) {
	old := testProject(t)
	updated := old
	updated.UpdatedAt = "2099-01-01T00:00:00Z"
	assert.Empty(t, DiffProjects(&old, updated))
}

func TestDiffTaskFieldUpdate(t *testing.T) {
	old := testProject(t)
	updated, tasks := forkTasks(old)

	require.NoError(t, tasks[0].SetStatus(domain.StatusCompleted))
	// RFC3339 second resolution: force a differing timestamp, as a real edit would.
	tasks[0].UpdatedAt = "2099-01-01T00:00:00Z"
	drafts := DiffProjects(&old, updated)

	require.Len(t, drafts, 1)
	d := drafts[0]
	assert.Equal(t, KindTaskUpdate, d.Kind)
	assert.Equal(t, tasks[0].ID, d.EntityID)
	assert.Equal(t, domain.StatusCompleted, d.Fields["status"])
	assert.Contains(t, d.Fields, "completion_date")
	assert.Contains(t, d.Fields, "updated_at")
	assert.NotContains(t, d.Fields, "title", "unchanged fields stay out")
}

func TestDiffClearedFieldIsExplicit(t *testing.T) {
	old := testProject(t)
	require.NoError(t, old.Categories[0].Tasks[0].SetPriority(domain.PriorityHigh))
	updated, tasks := forkTasks(old)
	require.NoError(t, tasks[0].SetPriority(""))

	drafts := DiffProjects(&old, updated)
	require.Len(t, drafts, 1)
	priority, present := drafts[0].Fields["priority"]
	assert.True(t, present)
	assert.Equal(t, "", priority)
}

func TestDiffTaskDeleteIsTombstone(t *testing.T) {
	old := testProject(t)
	updated, tasks := forkTasks(old)
	deleted := tasks[0]
	updated.Categories[0].Tasks = tasks[1:]

	drafts := DiffProjects(&old, updated)
	assert.Equal(t, []string{KindTaskDelete, KindCategoryReorder}, kinds(drafts))
	assert.Equal(t, deleted.ID, drafts[0].EntityID)
	assert.Empty(t, drafts[0].Fields, "tombstones carry no fields")
}

func TestDiffCrossCategoryMove(t *testing.T) {
	old := testProject(t)
	other, err := domain.NewCategory("Later")
	require.NoError(t, err)
	old.Categories = append(old.Categories, other)

	updated, tasks := forkTasks(old)
	moved := tasks[0]
	updated.Categories[0].Tasks = tasks[1:]
	updated.Categories[1].Tasks = []domain.Task{moved}

	drafts := DiffProjects(&old, updated)
	assert.Equal(t, []string{KindTaskMove, KindCategoryReorder, KindCategoryReorder}, kinds(drafts))
	assert.Equal(t, moved.ID, drafts[0].EntityID)
	assert.Equal(t, other.ID, drafts[0].Fields["category_id"])
}

func TestDiffReorderWithinCategory(t *testing.T) {
	old := testProject(t)
	updated, tasks := forkTasks(old)
	a, b := tasks[0], tasks[1]
	updated.Categories[0].Tasks = []domain.Task{b, a}

	drafts := DiffProjects(&old, updated)
	require.Len(t, drafts, 1)
	assert.Equal(t, KindCategoryReorder, drafts[0].Kind)
	assert.Equal(t, []string{b.ID, a.ID}, drafts[0].Fields["task_ids"])
}

func TestDiffProjectRename(t *testing.T) {
	old := testProject(t)
	updated := old
	updated.Name = "Renamed"
	drafts := DiffProjects(&old, updated)
	require.Len(t, drafts, 1)
	assert.Equal(t, KindProjectUpdate, drafts[0].Kind)
	assert.Equal(t, "Renamed", drafts[0].Fields["name"])
}

func TestDiffCategoryDelete(t *testing.T) {
	old := testProject(t)
	updated := old
	updated.Categories = []domain.Category{}

	drafts := DiffProjects(&old, updated)
	assert.Equal(t, []string{
		KindTaskDelete, KindTaskDelete, KindCategoryDelete, KindProjectReorder,
	}, kinds(drafts))
	assert.Equal(t, old.Categories[0].ID, drafts[2].EntityID)
	assert.Equal(t, []string{}, drafts[3].Fields["category_ids"])
}

func TestDiffSeparatorCarriesKind(t *testing.T) {
	old := testProject(t)
	sep, err := domain.NewSeparator()
	require.NoError(t, err)
	updated, tasks := forkTasks(old)
	updated.Categories[0].Tasks = append([]domain.Task{sep}, tasks...)

	drafts := DiffProjects(&old, updated)
	assert.Equal(t, []string{KindTaskCreate, KindCategoryReorder}, kinds(drafts))
	assert.Equal(t, domain.KindSeparator, drafts[0].Fields["kind"])
	assert.NotContains(t, drafts[0].Fields, "status")
}
