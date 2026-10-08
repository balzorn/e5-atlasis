package authorization

type ResourceType string

const (
	ResourceInformationAsset ResourceType = "information_asset"
	ResourceChangeRequest    ResourceType = "change_request"
	ResourceApproval         ResourceType = "approval"
)

type Resource struct {
	Type ResourceType
	ID   string
}
