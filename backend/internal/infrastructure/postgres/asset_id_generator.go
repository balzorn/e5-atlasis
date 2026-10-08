package postgres

import (
	"context"
	"fmt"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const assetIDSequenceName = "information_asset_id_seq"

type AssetIDGenerator struct {
	db *DB
}

func NewAssetIDGenerator(db *DB) *AssetIDGenerator {
	return &AssetIDGenerator{db: db}
}

func (g *AssetIDGenerator) Next(ctx context.Context) (domainasset.AssetID, error) {
	var number int64
	if err := g.db.pool.QueryRow(
		ctx,
		"SELECT nextval('"+assetIDSequenceName+"')",
	).Scan(&number); err != nil {
		return "", fmt.Errorf("allocate information asset ID: %w", err)
	}

	if number < 1 || number > 99999 {
		return "", fmt.Errorf("allocated information asset ID %d is outside supported range", number)
	}

	id, err := domainasset.ParseAssetID(fmt.Sprintf("IA%05d", number))
	if err != nil {
		return "", fmt.Errorf("parse allocated information asset ID: %w", err)
	}

	return id, nil
}

var _ ports.AssetIDGenerator = (*AssetIDGenerator)(nil)
