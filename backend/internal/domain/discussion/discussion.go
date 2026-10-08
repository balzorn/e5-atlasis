package discussion

import (
	"fmt"
	"time"
)

type ThreadID string
type CommentID string
type ChangeID string

type ThreadStatus string

const (
	ThreadStatusOpen     ThreadStatus = "OPEN"
	ThreadStatusResolved ThreadStatus = "RESOLVED"
)

type Thread struct {
	ID              ThreadID
	ChangeRequestID string
	ChangeID        *ChangeID
	Status          ThreadStatus
	CreatedBy       string
	CreatedAt       time.Time
	ResolvedAt      *time.Time
	ResolvedBy      *string
}

type Comment struct {
	ID        CommentID
	ThreadID  ThreadID
	AuthorID  string
	Body      string
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

func (t Thread) CanResolve() bool {
	return t.Status == ThreadStatusOpen
}

func (t Thread) CanReopen() bool {
	return t.Status == ThreadStatusResolved
}

func (t Thread) CanComment() bool {
	return t.Status == ThreadStatusOpen
}

func (c Comment) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("comment ID is required")
	}

	if c.ThreadID == "" {
		return fmt.Errorf("thread ID is required")
	}

	if c.AuthorID == "" {
		return fmt.Errorf("author ID is required")
	}

	if c.Body == "" {
		return fmt.Errorf("comment body is required")
	}

	return nil
}
