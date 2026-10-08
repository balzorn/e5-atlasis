package change

import (
	"fmt"
	"time"
)

type ID string

type Status string

const (
	StatusDraft            Status = "DRAFT"
	StatusSubmitted        Status = "SUBMITTED"
	StatusUnderReview      Status = "UNDER_REVIEW"
	StatusChangesRequested Status = "CHANGES_REQUESTED"
	StatusApproved         Status = "APPROVED"
	StatusApplying         Status = "APPLYING"
	StatusApplied          Status = "APPLIED"
	StatusRejected         Status = "REJECTED"
	StatusCancelled        Status = "CANCELLED"
)

type FieldChange struct {
	ID       string
	Field    string
	OldValue any
	NewValue any
}

type ChangeRequest struct {
	ID          ID
	AssetID     string
	BaseVersion int
	Status      Status
	Initiator   string
	Title       string
	Description string
	Changes     []FieldChange
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c ChangeRequest) CanTransitionTo(target Status) bool {
	switch c.Status {
	case StatusDraft:
		return target == StatusSubmitted || target == StatusCancelled

	case StatusSubmitted:
		return target == StatusUnderReview ||
			target == StatusCancelled

	case StatusUnderReview:
		return target == StatusChangesRequested ||
			target == StatusApproved ||
			target == StatusRejected

	case StatusChangesRequested:
		return target == StatusUnderReview ||
			target == StatusCancelled

	case StatusApproved:
		return target == StatusApplying

	case StatusApplying:
		return target == StatusApplied

	default:
		return false
	}
}

func (c ChangeRequest) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("change request ID is required")
	}

	if c.AssetID == "" {
		return fmt.Errorf("asset ID is required")
	}

	if c.BaseVersion < 1 {
		return fmt.Errorf("base version must be greater than zero")
	}

	if c.Initiator == "" {
		return fmt.Errorf("initiator is required")
	}

	if c.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(c.Changes) == 0 {
		return fmt.Errorf("at least one change is required")
	}

	return nil
}
