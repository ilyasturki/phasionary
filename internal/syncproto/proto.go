package syncproto

import (
	"phasionary/internal/domain"
	"phasionary/internal/journal"
)

const (
	EnrollPath = "/v1/enroll"
	SyncPath   = "/v1/sync"
)

type EnrollRequest struct {
	Code     string `json:"code"`
	DeviceID string `json:"device_id"`
	Name     string `json:"name,omitempty"`
}

type EnrollResponse struct {
	Token string `json:"token"`
}

type SyncRequest struct {
	DeviceID string       `json:"device_id"`
	Cursor   uint64       `json:"cursor"`
	Ops      []journal.Op `json:"ops"`
}

type SyncResponse struct {
	AckedThroughSeq uint64            `json:"acked_through_seq"`
	ServerCursor    uint64            `json:"server_cursor"`
	Projects        []ProjectSnapshot `json:"projects"`
}

// Project is nil when Deleted is set.
type ProjectSnapshot struct {
	ID      string          `json:"id"`
	Deleted bool            `json:"deleted,omitempty"`
	Project *domain.Project `json:"project,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
