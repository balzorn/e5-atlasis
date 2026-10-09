package openapicontract

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestOpenAPIContract(t *testing.T) {
	specPath := repositoryFile("docs", "api", "openapi.yaml")

	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read OpenAPI spec: %v", err)
	}

	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatalf("invalid YAML: %v", err)
	}

	if got, _ := document["openapi"].(string); got != "3.1.0" {
		t.Fatalf("OpenAPI version = %q, want 3.1.0", got)
	}

	paths := getMap(t, document, "paths")
	components := getMap(t, document, "components")
	schemas := getMap(t, components, "schemas")
	parameters := getMap(t, components, "parameters")
	responses := getMap(t, components, "responses")

	expectedOperations := map[string][]string{
		"/healthz":                           {"get"},
		"/assets":                            {"post"},
		"/assets/{assetId}":                  {"get"},
		"/change-requests":                   {"post"},
		"/change-requests/{changeRequestId}": {"get"},
		"/change-requests/{changeRequestId}/submit":          {"post"},
		"/change-requests/{changeRequestId}/review":          {"post"},
		"/change-requests/{changeRequestId}/request-changes": {"post"},
		"/change-requests/{changeRequestId}/reject":          {"post"},
		"/change-requests/{changeRequestId}/approve":         {"post"},
		"/change-requests/{changeRequestId}/apply":           {"post"},
		"/change-requests/{changeRequestId}/approvals":       {"get", "post"},
		"/approvals/{approvalId}/approve":                    {"post"},
		"/approvals/{approvalId}/reject":                     {"post"},
	}

	for path, methods := range expectedOperations {
		pathValue, ok := paths[path]
		if !ok {
			t.Errorf("missing OpenAPI path %q", path)
			continue
		}

		pathItem, ok := pathValue.(map[string]any)
		if !ok {
			t.Errorf("path %q is not a mapping", path)
			continue
		}

		for _, method := range methods {
			operationValue, ok := pathItem[method]
			if !ok {
				t.Errorf("missing OpenAPI operation %s %s", strings.ToUpper(method), path)
				continue
			}

			if path != "/healthz" {
				operation, ok := operationValue.(map[string]any)
				if !ok {
					t.Errorf("operation %s %s is not a mapping", strings.ToUpper(method), path)
					continue
				}

				hasActorID := false
				if parameters, ok := operation["parameters"].([]any); ok {
					for _, parameterValue := range parameters {
						parameter, ok := parameterValue.(map[string]any)
						if ok && parameter["$ref"] == "#/components/parameters/ActorId" {
							hasActorID = true
							break
						}
					}
				}
				if !hasActorID {
					t.Errorf("operation %s %s does not document the principal header", strings.ToUpper(method), path)
				}

				operationResponses, ok := operation["responses"].(map[string]any)
				if !ok {
					t.Errorf("operation %s %s has no responses mapping", strings.ToUpper(method), path)
				} else if _, ok := operationResponses["401"]; !ok {
					t.Errorf("operation %s %s does not document HTTP 401", strings.ToUpper(method), path)
				}
			}
		}
	}

	requiredSchemas := []string{
		"Error",
		"InformationAsset",
		"SecurityProfile",
		"CreateInformationAssetRequest",
		"CreateChangeRequestRequest",
		"ChangeProposal",
		"ChangeRequest",
		"FieldChange",
		"FieldValue",
		"CreateApprovalRequest",
		"ApprovalDecisionRequest",
		"Approval",
		"ApprovalType",
		"ApprovalStatus",
	}

	for _, name := range requiredSchemas {
		if _, ok := schemas[name]; !ok {
			t.Errorf("missing OpenAPI schema %q", name)
		}
	}

	requiredParameters := []string{
		"ActorId",
		"AssetId",
		"ChangeRequestId",
		"ApprovalId",
	}

	for _, name := range requiredParameters {
		if _, ok := parameters[name]; !ok {
			t.Errorf("missing OpenAPI parameter %q", name)
		}
	}

	requiredResponses := []string{
		"Unauthorized",
		"BadRequest",
		"UnsupportedMediaType",
		"Conflict",
		"Forbidden",
		"NotFound",
		"InternalError",
	}

	for _, name := range requiredResponses {
		if _, ok := responses[name]; !ok {
			t.Errorf("missing OpenAPI response %q", name)
		}
	}

	refCount := 0
	walkRefs(document, func(ref string) {
		refCount++

		if !strings.HasPrefix(ref, "#/") {
			t.Errorf("unsupported non-local $ref %q", ref)
			return
		}

		if err := resolveLocalRef(document, ref); err != nil {
			t.Errorf("broken $ref %q: %v", ref, err)
		}
	})

	if refCount == 0 {
		t.Error("OpenAPI document contains no $ref entries")
	}

	t.Logf(
		"OpenAPI contract OK: version=3.1.0 paths=%d schemas=%d parameters=%d responses=%d refs=%d",
		len(paths),
		len(schemas),
		len(parameters),
		len(responses),
		refCount,
	)
}

func repositoryFile(parts ...string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../.."))
	return filepath.Join(append([]string{root}, parts...)...)
}

func getMap(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()

	value, ok := parent[key]
	if !ok {
		t.Fatalf("missing top-level mapping %q", key)
	}

	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("mapping %q has unexpected type %T", key, value)
	}

	return result
}

func walkRefs(value any, visit func(string)) {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if key == "$ref" {
				if ref, ok := child.(string); ok {
					visit(ref)
				}
				continue
			}

			walkRefs(child, visit)
		}

	case []any:
		for _, child := range current {
			walkRefs(child, visit)
		}
	}
}

func resolveLocalRef(document map[string]any, ref string) error {
	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")

	var current any = document

	for _, part := range parts {
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		mapping, ok := current.(map[string]any)
		if !ok {
			return fmt.Errorf("expected mapping before %q", part)
		}

		next, ok := mapping[part]
		if !ok {
			return fmt.Errorf("target %q not found", part)
		}

		current = next
	}

	return nil
}
