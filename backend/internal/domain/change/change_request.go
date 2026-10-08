package change

import (
	"fmt"
	"regexp"
	"time"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

type ID string

var changeRequestIDPattern = regexp.MustCompile(`^CR[0-9]{5}package change

import (
	"fmt"
	"regexp"
	"time"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

)

func ParseChangeRequestID(value string) (ID, error) {
	if !changeRequestIDPattern.MatchString(value) {
		return "", fmt.Errorf("invalid change request ID")
	}

	return ID(value), nil
}

type Status string

func (id ID) String() string {
	return string(id)
}

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
	Field    domainasset.FieldName
	OldValue domainasset.FieldValue
	NewValue domainasset.FieldValue
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

func (c *ChangeRequest) TransitionTo(target Status) error {
	if !c.CanTransitionTo(target) {
		return fmt.Errorf("invalid change request transition: %s -> %s", c.Status, target)
	}

	c.Status = target
	return nil
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
	if _, err := ParseChangeRequestID(c.ID.String()); err != nil {
		return err
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
