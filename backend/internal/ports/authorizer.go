package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/authorization"
)

type Authorizer interface {
	Authorize(
		ctx context.Context,
		subject authorization.Subject,
		action authorization.Action,
		resource authorization.Resource,
	) error
}
