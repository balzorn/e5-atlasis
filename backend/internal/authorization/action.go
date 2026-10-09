package authorization

// Action is an application-level action, independent of HTTP method or route.
type Action string

const (
	ActionInformationAssetRead   Action = "information_asset.read"
	ActionInformationAssetCreate Action = "information_asset.create"

	ActionChangeRequestRead           Action = "change_request.read"
	ActionChangeRequestCreate         Action = "change_request.create"
	ActionChangeRequestSubmit         Action = "change_request.submit"
	ActionChangeRequestReview         Action = "change_request.review"
	ActionChangeRequestRequestChanges Action = "change_request.request_changes"
	ActionChangeRequestReject         Action = "change_request.reject"
	ActionChangeRequestApprove        Action = "change_request.approve"
	ActionChangeRequestApply          Action = "change_request.apply"

	ActionApprovalRead    Action = "approval.read"
	ActionApprovalCreate  Action = "approval.create"
	ActionApprovalApprove Action = "approval.approve"
	ActionApprovalReject  Action = "approval.reject"
)

// IsKnown reports whether a is an action recognized by this application.
func (a Action) IsKnown() bool {
	switch a {
	case ActionInformationAssetRead,
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
		ActionApprovalReject:
		return true
	default:
		return false
	}
}
