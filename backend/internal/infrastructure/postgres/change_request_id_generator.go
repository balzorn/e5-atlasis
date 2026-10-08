package postgres

import (
	"context"
	"fmt"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const changeRequestIDSequenceName = "change_request_id_seq"

type ChangeRequestIDGenerator struct {
	db *DB
}

func NewChangeRequestIDGenerator(db *DB) *ChangeRequestIDGenerator {
	return &ChangeRequestIDGenerator{db: db}
}

func (g *ChangeRequestIDGenerator) Next(ctx context.Context) (domainchange.ID, error) {
	var number int64
	if err := g.db.pool.QueryRow(
		ctx,
		"SELECT nextval('"+changeRequestIDSequenceName+"')",
	).Scan(&number); err != nil {
		return "", fmt.Errorf("allocate change request ID: %w", err)
	}

	if number < 1 || number > 99999 {
		return "", fmt.Errorf("allocated change request ID %d is outside supported range", number)
	}

	return domainchange.ID(fmt.Sprintf("CR%05d", number)), nil
}

var _ ports.ChangeRequestIDGenerator = (*ChangeRequestIDGenerator)(nil)
