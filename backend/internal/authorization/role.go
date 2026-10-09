package authorization

// Role is a logical role used by authorization policies.
// Roles are assigned by the trusted identity/context resolver, never by request payloads.
type Role string

const (
	RoleInitiator       Role = "Initiator"
	RoleReviewer        Role = "Reviewer"
	RoleApprover        Role = "Approver"
	RoleAssetOwner      Role = "Asset Owner"
	RoleSecurityOfficer Role = "Security Officer"
	RoleAdministrator   Role = "Administrator"
)

// IsKnown reports whether r is one of the roles recognized by this application.
func (r Role) IsKnown() bool {
	switch r {
	case RoleInitiator,
		RoleReviewer,
		RoleApprover,
		RoleAssetOwner,
		RoleSecurityOfficer,
		RoleAdministrator:
		return true
	default:
		return false
	}
}
