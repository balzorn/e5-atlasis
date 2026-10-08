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

func (a Approval) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("approval ID is required")
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
