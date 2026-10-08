package approval

import (
	"fmt"
	"time"
)

type ID string
type Type string
type Status string

const (
	TypeAssetOwner   Type = "ASSET_OWNER"
	TypeSecurity     Type = "SECURITY"
	TypeOrganization Type = "ORGANIZATION"
	TypeSystemOwner  Type = "SYSTEM_OWNER"
)

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
)

type Approval struct {
	ID              ID
	ChangeRequestID string
	Type            Type
	Status          Status
	Required        bool
	ApproverID      string
	DecidedAt       *time.Time
	Comment         string
}

func (a *Approval) Approve(decidedBy string, comment string, decidedAt time.Time) error {
	if a.Status != StatusPending {
		return fmt.Errorf("approval %q is not pending", a.ID)
	}

	if decidedBy == "" {
		return fmt.Errorf("decided by is required")
	}

	if a.ApproverID != decidedBy {
		return fmt.Errorf("user %q is not assigned as approver", decidedBy)
	}

	a.Status = StatusApproved
	a.DecidedAt = &decidedAt
	a.Comment = comment

	return nil
}

func (a *Approval) Reject(decidedBy string, comment string, decidedAt time.Time) error {
	if a.Status != StatusPending {
		return fmt.Errorf("approval %q is not pending", a.ID)
	}

	if decidedBy == "" {
		return fmt.Errorf("decided by is required")
	}

	if a.ApproverID != decidedBy {
		return fmt.Errorf("user %q is not assigned as approver", decidedBy)
	}

	a.Status = StatusRejected
	a.DecidedAt = &decidedAt
	a.Comment = comment

	return nil
}

func (a Approval) Validate() error {
	if _, err := ParseApprovalID(a.ID.String()); err != nil {
		return err
	}

	if a.ChangeRequestID == "" {
		return fmt.Errorf("change request ID is required")
	}

	if a.ApproverID == "" {
		return fmt.Errorf("approver ID is required")
	}

	switch a.Type {
	case TypeAssetOwner,
		TypeSecurity,
		TypeOrganization,
		TypeSystemOwner:
	default:
		return fmt.Errorf("invalid approval type %q", a.Type)
	}

	switch a.Status {
	case StatusPending,
		StatusApproved,
		StatusRejected:
	default:
		return fmt.Errorf("invalid approval status %q", a.Status)
	}

	return nil
}
