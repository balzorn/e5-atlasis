package authorization

// Subject is the authenticated actor presented to the authorization boundary.
//
// ID is sourced from the authenticated principal. Roles and OrganizationIDs
// must be populated only from trusted identity claims or server-side mappings;
// callers must not populate them directly from untrusted request data.
type Subject struct {
	ID              string
	Roles           []Role
	OrganizationIDs []string
}
