package authorization

import "testing"

func TestKnownRoles(t *testing.T) {
	roles := []Role{
		RoleInitiator,
		RoleReviewer,
		RoleApprover,
		RoleAssetOwner,
		RoleSecurityOfficer,
		RoleAdministrator,
	}

	for _, role := range roles {
		if !role.IsKnown() {
			t.Errorf("IsKnown(%q) = false, want true", role)
		}
	}

	if (Role("Superuser")).IsKnown() {
		t.Error("unknown role must not be recognized")
	}
}

func TestKnownActions(t *testing.T) {
	actions := []Action{
		ActionInformationAssetRead,
		ActionInformationAssetCreate,
		ActionChangeRequestRead,
		ActionChangeRequestCreate,
		ActionChangeRequestSubmit,
		ActionChangeRequestReview,
		ActionChangeRequestRequestChanges,
		ActionChangeRequestReject,
		ActionChangeRequestApprove,
		ActionChangeRequestApply,
		ActionApprovalRead,
		ActionApprovalCreate,
		ActionApprovalApprove,
		ActionApprovalReject,
	}

	for _, action := range actions {
		if !action.IsKnown() {
			t.Errorf("IsKnown(%q) = false, want true", action)
		}
	}

	if (Action("change_request.force_apply")).IsKnown() {
		t.Error("unknown action must not be recognized")
	}
}

func TestKnownResourceTypes(t *testing.T) {
	resourceTypes := []ResourceType{
		ResourceInformationAsset,
		ResourceChangeRequest,
		ResourceApproval,
	}

	for _, resourceType := range resourceTypes {
		if !resourceType.IsKnown() {
			t.Errorf("IsKnown(%q) = false, want true", resourceType)
		}
	}

	if (ResourceType("unknown")).IsKnown() {
		t.Error("unknown resource type must not be recognized")
	}
}
