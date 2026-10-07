# E5-ATLASIS

Information Security Asset & Infrastructure System for 5 Element.

## Architecture
- Clean Architecture + Hexagonal Architecture + DDD
- Backend: Go
- Frontend: Vue 3 + TypeScript + Vite
- API: OpenAPI-first
- PostgreSQL
- Authorization: Cedar
- Compliance: OPA/Rego
- Observability: OpenTelemetry
- Deployment: Kubernetes

## Rules
- Domain must not depend on infrastructure.
- Use pgx/v5; no ORM.
- OpenAPI is the source of truth for the API.
- Controlled asset changes use Change Requests.
- Authorization is default-deny and policy-driven.
- Never commit secrets.
- Never log secrets, credentials or tokens.
- Preserve auditability and immutable history.
- Use context.Context.
- Add tests for behavior changes.

## Secure Go
For Go code, follow:
.opencode/skills/secure-go/SKILL.md

## AI agents
Use specialized agents when appropriate:
- software-architect
- backend-architect
- security-architect
- code-reviewer
- database-optimizer
- frontend-developer
- devops-automator
- api-tester
- reality-checker

## Git
- Work on feature branches.
- Never modify main directly.
- Keep commits small and focused.
- Review diffs before committing.
