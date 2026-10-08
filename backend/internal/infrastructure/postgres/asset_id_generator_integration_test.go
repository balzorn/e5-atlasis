package postgres

import (
	"context"
	"testing"
	"time"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

func TestPostgreSQLAssetIDGenerator(t *testing.T) {
	db := newIntegrationDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	generator := NewAssetIDGenerator(db)

	first, err := generator.Next(ctx)
	if err != nil {
		t.Fatalf("first ID: %v", err)
	}

	second, err := generator.Next(ctx)
	if err != nil {
		t.Fatalf("second ID: %v", err)
	}

	if _, err := domainasset.ParseAssetID(first.String()); err != nil {
		t.Fatalf("first generated ID %q is invalid: %v", first, err)
	}
	if _, err := domainasset.ParseAssetID(second.String()); err != nil {
		t.Fatalf("second generated ID %q is invalid: %v", second, err)
	}

	if first == second {
		t.Fatalf("generated IDs must be unique: %q", first)
	}

	wantNext := first
	if second <= first {
		t.Fatalf("generated IDs are not monotonic: first=%q second=%q", first, second)
	}
	_ = wantNext
}
