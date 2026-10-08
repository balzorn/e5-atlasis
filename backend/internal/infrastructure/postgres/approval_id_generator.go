package postgres

import (
	"context"
	"fmt"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const approvalIDSequenceName = "approval_id_seq"

type ApprovalIDGenerator struct {
	db *DB
}

func NewApprovalIDGenerator(db *DB) *ApprovalIDGenerator {
	return &ApprovalIDGenerator{db: db}
}

func (g *ApprovalIDGenerator) Next(ctx context.Context) (domainapproval.ID, error) {
	var number int64
	if err := g.db.pool.QueryRow(
		ctx,
		"SELECT nextval('"+approvalIDSequenceName+"')",
	).Scan(&number); err != nil {
		return "", fmt.Errorf("allocate approval ID: %w", err)
	}

	if number < 1 || number > 99999 {
		return "", fmt.Errorf("allocated approval ID %d is outside supported range", number)
	}

	id, err := domainapproval.ParseApprovalID(fmt.Sprintf("APR%05d", number))
	if err != nil {
		return "", fmt.Errorf("parse allocated approval ID: %w", err)
	}

	return id, nil
}

var _ ports.ApprovalIDGenerator = (*ApprovalIDGenerator)(nil)
