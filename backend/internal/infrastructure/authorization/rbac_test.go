package authorization

import (
	"context"
	"errors"
	"testing"

	authz "github.com/balzorn/e5-atlasis/backend/internal/authorization"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const testOrganizationID = "ORG001"

func TestRBACAuthorizer(t *testing.T) {
	tests := []struct {
		name     string
		subject  authz.Subject
		action   authz.Action
		resource authz.Resource
		wantErr  error
	}{
		{
			name: "initiator creates an asset in an assigned organization",
			subject: testSubject("USR001", authz.RoleInitiator),
			action: authz.ActionInformationAssetCreate,
			resource: authz.Resource{
				Type: authz.ResourceInformationAsset, OrganizationID: testOrganizationID,
			},
		},
		{
			name: "initiator submits own change request",
			subject: testSubject("USR001", authz.RoleInitiator),
			action: authz.ActionChangeRequestSubmit,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID, InitiatorID: "USR001",
			},
		},
		{
			name: "initiator cannot submit another user's change request",
			subject: testSubject("USR001", authz.RoleInitiator),
			action: authz.ActionChangeRequestSubmit,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID, InitiatorID: "USR002",
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "reviewer can review a change request in scope",
			subject: testSubject("USR003", authz.RoleReviewer),
			action: authz.ActionChangeRequestReview,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID, InitiatorID: "USR001",
			},
		},
		{
			name: "security officer cannot reject a change request",
			subject: testSubject("USR003", authz.RoleSecurityOfficer),
			action: authz.ActionChangeRequestReject,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID, InitiatorID: "USR001",
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "assigned approver can approve assigned approval",
			subject: testSubject("USR002", authz.RoleApprover),
			action: authz.ActionApprovalApprove,
			resource: authz.Resource{
				Type: authz.ResourceApproval, ID: "APR00001",
				OrganizationID: testOrganizationID, ApproverID: "USR002",
			},
		},
		{
			name: "approver cannot decide approval assigned to another user",
			subject: testSubject("USR002", authz.RoleApprover),
			action: authz.ActionApprovalApprove,
			resource: authz.Resource{
				Type: authz.ResourceApproval, ID: "APR00001",
				OrganizationID: testOrganizationID, ApproverID: "USR003",
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "reviewer cannot decide approval without approver role",
			subject: testSubject("USR002", authz.RoleReviewer),
			action: authz.ActionApprovalReject,
			resource: authz.Resource{
				Type: authz.ResourceApproval, ID: "APR00001",
				OrganizationID: testOrganizationID, ApproverID: "USR002",
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "asset owner can read owned asset",
			subject: testSubject("USR001", authz.RoleAssetOwner),
			action: authz.ActionInformationAssetRead,
			resource: authz.Resource{
				Type: authz.ResourceInformationAsset, ID: "IA00001",
				OrganizationID: testOrganizationID, OwnerID: "USR001",
			},
		},
		{
			name: "asset owner cannot read another user's asset",
			subject: testSubject("USR001", authz.RoleAssetOwner),
			action: authz.ActionInformationAssetRead,
			resource: authz.Resource{
				Type: authz.ResourceInformationAsset, ID: "IA00001",
				OrganizationID: testOrganizationID, OwnerID: "USR002",
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "subject outside organization is denied",
			subject: authz.Subject{
				ID: "USR001", Roles: []authz.Role{authz.RoleReviewer},
				OrganizationIDs: []string{"ORG002"},
			},
			action: authz.ActionChangeRequestReview,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID,
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "organization scope is required on resource context",
			subject: testSubject("USR001", authz.RoleReviewer),
			action: authz.ActionChangeRequestReview,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "unknown role denies entire subject",
			subject: authz.Subject{
				ID: "USR001", Roles: []authz.Role{authz.RoleReviewer, "CustomAdmin"},
				OrganizationIDs: []string{testOrganizationID},
			},
			action: authz.ActionChangeRequestReview,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID,
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "administrator has no implicit bypass",
			subject: testSubject("USR999", authz.RoleAdministrator),
			action: authz.ActionChangeRequestReject,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID,
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "change request application remains denied until explicitly granted",
			subject: testSubject("USR003", authz.RoleReviewer, authz.RoleAdministrator),
			action: authz.ActionChangeRequestApply,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID,
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "unknown action is denied",
			subject: testSubject("USR003", authz.RoleReviewer),
			action: authz.Action("change_request.purge"),
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID,
			},
			wantErr: ports.ErrForbidden,
		},
		{
			name: "action on wrong resource type is denied",
			subject: testSubject("USR003", authz.RoleReviewer),
			action: authz.ActionApprovalApprove,
			resource: authz.Resource{
				Type: authz.ResourceChangeRequest, ID: "CR00001",
				OrganizationID: testOrganizationID,
			},
			wantErr: ports.ErrForbidden,
		},
	}

	authorizer := NewRBACAuthorizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authorizer.Authorize(context.Background(), tt.subject, tt.action, tt.resource)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Authorize() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Authorize() error = %v, want error wrapping %v", err, tt.wantErr)
			}
		})
	}
}

func TestRBACAuthorizerHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewRBACAuthorizer().Authorize(
		ctx,
		testSubject("USR003", authz.RoleReviewer),
		authz.ActionChangeRequestReview,
		authz.Resource{
			Type: authz.ResourceChangeRequest, ID: "CR00001",
			OrganizationID: testOrganizationID,
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Authorize() error = %v, want context.Canceled", err)
	}
}

func testSubject(id string, roles ...authz.Role) authz.Subject {
	return authz.Subject{
		ID:              id,
		Roles:           roles,
		OrganizationIDs: []string{testOrganizationID},
	}
}
