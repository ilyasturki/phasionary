package journal

// Wire values, persisted in the journal: renaming one invalidates pending ops.
// Separators travel as task kinds; order travels only in reorder kinds.
const (
	KindProjectCreate  = "project.create"
	KindProjectUpdate  = "project.update"
	KindProjectDelete  = "project.delete"
	KindProjectReorder = "project.reorder"

	KindCategoryCreate  = "category.create"
	KindCategoryUpdate  = "category.update"
	KindCategoryDelete  = "category.delete"
	KindCategoryReorder = "category.reorder"

	KindTaskCreate = "task.create"
	KindTaskUpdate = "task.update"
	KindTaskDelete = "task.delete"
	KindTaskMove   = "task.move"
)

type Op struct {
	OpID      string `json:"op_id"`
	DeviceID  string `json:"device_id"`
	Seq       uint64 `json:"seq"`
	Timestamp string `json:"ts"`
	Draft
}

// Fields keys are the domain JSON tags. On an update a present empty value is
// a deliberate clear, an absent key an untouched field; deletes carry none.
type Draft struct {
	Kind      string         `json:"kind"`
	ProjectID string         `json:"project_id"`
	EntityID  string         `json:"entity_id,omitempty"`
	Fields    map[string]any `json:"fields,omitempty"`
}
