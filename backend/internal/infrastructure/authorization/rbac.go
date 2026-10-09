package authorization

import (
	"context"
	"fmt"
	"strings"

	authz "github.com/balzorn/e5-atlasis/backend/internal/authorization"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

// RBACAuthorizer evaluates the initial role and relationship policy.
//
// It intentionally denies change_request.apply until the application defines
// an explicit role grant for that sensitive operation. Administrator does not
// implicitly bypass the policy.
type RBACAuthorizer struct{}

func NewRBACAuthorizer() *RBACAuthorizer {
	return &RBACAuthorizer{}
}

func (a *RBACAuthorizer) Authorize(
	ctx context.Context,
	subject authz.Subject,
	action authz.Action,
	resource authz.Resource,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if !validSubject(subject) ||
		!action.IsKnown() ||
		!validResourceContext(action, resource) ||
		!subjectBelongsToOrganization(subject, resource.OrganizationID) {
		return deny()
	}

	if grants(subject, action, resource) {
		return nil
	}

	return deny()
}

func validSubject(subject authz.Subject) bool {
	if strings.TrimSpace(subject.ID) == "" || len(subject.Roles) == 0 {
		return false
	}

	for _, role := range subject.Roles {
		if !role.IsKnown() {
			return false
		}
	}

	for _, organizationID := range subject.OrganizationIDs {
		if strings.TrimSpace(organizationID) == "" {
			return false
		}
	}

	return true
}

func validResourceContext(action authz.Action, resource authz.Resource) bool {
	if !resource.Type.IsKnown() ||
		strings.TrimSpace(resource.OrganizationID) == "" ||
		!actionMatchesResource(action, resource.Type) {
		return false
	}

	isCreate := action == authz.ActionInformationAssetCreate ||
		action == authz.ActionChangeRequestCreate ||
		action == authz.ActionApprovalCreate

	if isCreate {
		if strings.TrimSpace(resource.ID) != "" {
			return false
		}
		switch resource.Type {
		case authz.ResourceInformationAsset:
			return strings.TrimSpace(resource.ParentID) == ""
		case authz.ResourceChangeRequest, authz.ResourceApproval:
			return strings.TrimSpace(resource.ParentID) != ""
		default:
			return false
		}
	}

	// Reading the approvals collection uses the parent Change Request as context.
	if action == authz.ActionApprovalRead &&
		strings.TrimSpace(resource.ID) == "" {
		return strings.TrimSpace(resource.ParentID) != ""
	}

	return strings.TrimSpace(resource.ID) != ""
}

func actionMatchesResource(action authz.Action, resourceType authz.ResourceType) bool {
	switch resourceType {
	case authz.ResourceInformationAsset:
		return action == authz.ActionInformationAssetRead ||
			action == authz.ActionInformationAssetCreate

	case authz.ResourceChangeRequest:
		switch action {
		case authz.ActionChangeRequestRead,
			authz.ActionChangeRequestCreate,
			authz.ActionChangeRequestSubmit,
			authz.ActionChangeRequestReview,
			authz.ActionChangeRequestRequestChanges,
			authz.ActionChangeRequestReject,
			authz.ActionChangeRequestApprove,
			authz.ActionChangeRequestApply:
			return true
		default:
			return false
		}

	case authz.ResourceApproval:
		switch action {
		case authz.ActionApprovalRead,
			authz.ActionApprovalCreate,
			authz.ActionApprovalApprove,
			authz.ActionApprovalReject:
			return true
		default:
			return false
		}

	default:
		return false
	}
}

func subjectBelongsToOrganization(subject authz.Subject, organizationID string) bool {
	for _, candidate := range subject.OrganizationIDs {
		if strings.TrimSpace(candidate) == strings.TrimSpace(organizationID) {
			return true
		}
	}
	return false
}

func grants(subject authz.Subject, action authz.Action, resource authz.Resource) bool {
	for _, role := range subject.Roles {
		switch action {
		case authz.ActionInformationAssetRead:
			switch role {
			case authz.RoleInitiator, authz.RoleReviewer, authz.RoleSecurityOfficer:
				return true
			case authz.RoleAssetOwner:
				if resource.OwnerID == subject.ID {
					return true
				}
			}

		case authz.ActionInformationAssetCreate:
			if role == authz.RoleInitiator {
				return true
			}

		case authz.ActionChangeRequestRead:
			switch role {
			case authz.RoleInitiator:
				if resource.InitiatorID == subject.ID {
					return true
				}
			case authz.RoleReviewer, authz.RoleSecurityOfficer:
				return true
			case authz.RoleAssetOwner:
				if resource.OwnerID == subject.ID {
					return true
				}
			}

		case authz.ActionChangeRequestCreate:
			if role == authz.RoleInitiator {
				return true
			}

		case authz.ActionChangeRequestSubmit:
			if role == authz.RoleInitiator && resource.InitiatorID == subject.ID {
				return true
			}

		case authz.ActionChangeRequestReview:
			if role == authz.RoleReviewer || role == authz.RoleSecurityOfficer {
				return true
			}

		case authz.ActionChangeRequestRequestChanges:
			if role == authz.RoleReviewer || role == authz.RoleSecurityOfficer {
				return true
			}

		case authz.ActionChangeRequestReject, authz.ActionChangeRequestApprove:
			if role == authz.RoleReviewer {
				return true
			}

		case authz.ActionChangeRequestApply:
			// Deliberately ungranted until the execution responsibility is defined.

		case authz.ActionApprovalRead:
			switch role {
			case authz.RoleReviewer, authz.RoleSecurityOfficer:
				return true
			case authz.RoleInitiator:
				if resource.InitiatorID == subject.ID {
					return true
				}
			case authz.RoleApprover:
				if resource.ApproverID == subject.ID {
					return true
				}
			case authz.RoleAssetOwner:
				if resource.OwnerID == subject.ID {
					return true
				}
			}

		case authz.ActionApprovalCreate:
			if role == authz.RoleReviewer || role == authz.RoleSecurityOfficer {
				return true
			}

		case authz.ActionApprovalApprove, authz.ActionApprovalReject:
			if role == authz.RoleApprover && resource.ApproverID == subject.ID {
				return true
			}
		}
	}

	return false
}

func deny() error {
	return fmt.Errorf("%w: authorization denied", ports.ErrForbidden)
}

var _ ports.Authorizer = (*RBACAuthorizer)(nil)
