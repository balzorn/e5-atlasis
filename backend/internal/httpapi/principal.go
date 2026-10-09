package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/balzorn/e5-atlasis/backend/internal/authorization"
)

var errNoAuthenticatedPrincipal = errors.New("no authenticated principal")

// PrincipalResolver derives a subject from a trusted authentication mechanism.
type PrincipalResolver interface {
	Resolve(r *http.Request) (authorization.Subject, error)
}

type rejectingPrincipalResolver struct{}

func (rejectingPrincipalResolver) Resolve(*http.Request) (authorization.Subject, error) {
	return authorization.Subject{}, errNoAuthenticatedPrincipal
}

// NewRejectingPrincipalResolver returns a resolver that rejects every API request.
// Use it until a trusted authentication mechanism has been configured.
func NewRejectingPrincipalResolver() PrincipalResolver {
	return rejectingPrincipalResolver{}
}

type developmentHeaderPrincipalResolver struct{}

// NewDevelopmentHeaderPrincipalResolver enables the legacy X-Actor-ID header for
// local development only. The server startup configuration must restrict binding
// to a loopback IP address before this resolver is used.
func NewDevelopmentHeaderPrincipalResolver() PrincipalResolver {
	return developmentHeaderPrincipalResolver{}
}

func (developmentHeaderPrincipalResolver) Resolve(r *http.Request) (authorization.Subject, error) {
	actorID := strings.TrimSpace(r.Header.Get("X-Actor-ID"))
	if actorID == "" || len(actorID) > maxActorIDLength {
		return authorization.Subject{}, errNoAuthenticatedPrincipal
	}

	// This development resolver intentionally supplies no roles or organization
	// claims. Never infer authorization privileges from an actor ID header.
	return authorization.Subject{ID: actorID}, nil
}

type principalContextKey struct{}

func principalFromContext(ctx context.Context) (authorization.Subject, bool) {
	subject, ok := ctx.Value(principalContextKey{}).(authorization.Subject)
	if !ok {
		return authorization.Subject{}, false
	}
	return subject, true
}

func (h *Handler) authenticateAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
			next.ServeHTTP(w, r)
			return
		}

		if h.principalResolver == nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "a valid authenticated principal is required")
			return
		}

		subject, err := h.principalResolver.Resolve(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "a valid authenticated principal is required")
			return
		}

		subject.ID = strings.TrimSpace(subject.ID)
		if subject.ID == "" || len(subject.ID) > maxActorIDLength {
			writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "a valid authenticated principal is required")
			return
		}

		ctx := context.WithValue(r.Context(), principalContextKey{}, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
