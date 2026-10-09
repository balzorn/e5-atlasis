package authorization

// ResourceType identifies a resource class understood by authorization policies.
type ResourceType string

const (
	ResourceInformationAsset ResourceType = "information_asset"
	ResourceChangeRequest    ResourceType = "change_request"
	ResourceApproval         ResourceType = "approval"
)

// IsKnown reports whether t is a resource type recognized by this application.
func (t ResourceType) IsKnown() bool {
	switch t {
	case ResourceInformationAsset, ResourceChangeRequest, ResourceApproval:
		return true
	default:
		return false
	}
}

// Resource is the policy-relevant context for an action.
//
// ID may be empty for create operations. ParentID identifies the containing
// resource when authorizing a collection or a not-yet-created child resource.
// Only attributes loaded or derived by the server may be supplied here.
type Resource struct {
	Type             ResourceType
	ID               string
	ParentID         string
	OrganizationID   string
	OwnerID          string
	InitiatorID      string
	ApproverID       string
	Status           string
}
