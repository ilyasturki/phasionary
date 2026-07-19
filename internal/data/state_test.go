package data

import (
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

func newLoadedStateManager(t *testing.T, dir, currentDir string) *StateManager {
	t.Helper()
	m := NewStateManager(dir, currentDir)
	if err := m.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	return m
}

// The TUI and `phasionary serve` each hold their own StateManager over the same
// file, and every save rewrites it whole. Without the reload-before-write, the
// second manager's save would drop the first manager's key.
func TestStateManagerConcurrentWritersDoNotClobber(t *testing.T) {
	dir := t.TempDir()
	tui := newLoadedStateManager(t, dir, "/work/project")
	serve := newLoadedStateManager(t, dir, "/work/project")

	if err := tui.SetProjectOrder([]string{"p1", "p2"}); err != nil {
		t.Fatalf("set order: %v", err)
	}
	// serve's cache predates the write above; its own save must not erase it.
	if err := serve.SetFoldedCategories("p1", []string{"cat-a"}); err != nil {
		t.Fatalf("set folds: %v", err)
	}

	fresh := newLoadedStateManager(t, dir, "/work/project")
	if got := fresh.GetProjectOrder(); !slices.Equal(got, []string{"p1", "p2"}) {
		t.Fatalf("project order clobbered: %v", got)
	}
	if got := fresh.GetFoldedCategories("p1"); !slices.Equal(got, []string{"cat-a"}) {
		t.Fatalf("folds clobbered: %v", got)
	}
}

// The phone folds a category while the TUI has the project open; the TUI reads
// folds at project open, so its next read must see the new value.
func TestGetFoldedCategoriesSeesOtherProcessWrites(t *testing.T) {
	dir := t.TempDir()
	tui := newLoadedStateManager(t, dir, "")
	serve := newLoadedStateManager(t, dir, "")

	if got := tui.GetFoldedCategories("p1"); len(got) != 0 {
		t.Fatalf("want no folds initially, got %v", got)
	}
	if err := serve.SetFoldedCategories("p1", []string{"cat-a", "cat-b"}); err != nil {
		t.Fatalf("set folds: %v", err)
	}

	got := tui.GetFoldedCategories("p1")
	if !slices.Equal(got, []string{"cat-a", "cat-b"}) {
		t.Fatalf("stale folds: want [cat-a cat-b], got %v", got)
	}
}

func TestSetFoldedCategoriesEmptyClearsProject(t *testing.T) {
	dir := t.TempDir()
	m := newLoadedStateManager(t, dir, "")

	if err := m.SetFoldedCategories("p1", []string{"cat-a"}); err != nil {
		t.Fatalf("set folds: %v", err)
	}
	if err := m.SetFoldedCategories("p1", nil); err != nil {
		t.Fatalf("clear folds: %v", err)
	}

	fresh := newLoadedStateManager(t, dir, "")
	if got := fresh.GetFoldedCategories("p1"); len(got) != 0 {
		t.Fatalf("want folds cleared, got %v", got)
	}
}

// A returned fold slice must not alias the cache, or a caller sorting it in
// place would silently rewrite state.
func TestGetFoldedCategoriesReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	m := newLoadedStateManager(t, dir, "")
	if err := m.SetFoldedCategories("p1", []string{"cat-a"}); err != nil {
		t.Fatalf("set folds: %v", err)
	}

	got := m.GetFoldedCategories("p1")
	got[0] = "mutated"

	if again := m.GetFoldedCategories("p1"); again[0] != "cat-a" {
		t.Fatalf("caller mutation leaked into state: %v", again)
	}
}

func TestUnlinkDirReturnsPreviousAndIsNoOpWhenUnlinked(t *testing.T) {
	dir := t.TempDir()
	m := newLoadedStateManager(t, dir, "/work/project")

	prev, err := m.UnlinkDir()
	if err != nil || prev != "" {
		t.Fatalf("unlink with no link: want (\"\", nil), got (%q, %v)", prev, err)
	}

	if err := m.SetProjectForDir("p1"); err != nil {
		t.Fatalf("link: %v", err)
	}
	prev, err = m.UnlinkDir()
	if err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if prev != "p1" {
		t.Fatalf("want previous id p1, got %q", prev)
	}
	if got := newLoadedStateManager(t, dir, "/work/project").GetProjectForDir(); got != "" {
		t.Fatalf("link survived unlink: %q", got)
	}
}

// The API server writes folds from concurrent request goroutines. Run with
// -race to make this meaningful.
func TestStateManagerConcurrentWritesAreSerialized(t *testing.T) {
	dir := t.TempDir()
	m := newLoadedStateManager(t, dir, "")

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := m.SetFoldedCategories("p1", []string{"cat"}); err != nil {
				t.Errorf("set folds: %v", err)
			}
			_ = m.GetFoldedCategories("p1")
		}(i)
	}
	wg.Wait()

	if got := newLoadedStateManager(t, dir, "").GetFoldedCategories("p1"); !slices.Equal(got, []string{"cat"}) {
		t.Fatalf("want [cat], got %v", got)
	}
}

func TestStateManagerPathIsStateJSON(t *testing.T) {
	dir := t.TempDir()
	m := NewStateManager(dir, "")
	if want := filepath.Join(dir, "state.json"); m.path != want {
		t.Fatalf("path: want %q, got %q", want, m.path)
	}
}
