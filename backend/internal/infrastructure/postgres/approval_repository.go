package postgres

import (
	"context"
	"fmt"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type ApprovalRepository struct {
	db *DB
}

func NewApprovalRepository(db *DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

func (r *ApprovalRepository) Create(
	ctx context.Context,
	a domainapproval.Approval,
) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("validate approval: %w", err)
	}

	_, err := r.db.pool.Exec(
		ctx,
		`
		INSERT INTO approvals (
			id,
			change_request_id,
			type,
			status,
			required,
			approver_id,
			decided_at,
			comment
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		string(a.ID),
		a.ChangeRequestID,
		a.Type,
		a.Status,
		a.Required,
		a.ApproverID,
		a.DecidedAt,
		a.Comment,
	)
	if err != nil {
		return fmt.Errorf("insert approval: %w", err)
	}

	return nil
}

func (r *ApprovalRepository) GetByID(
	ctx context.Context,
	id domainapproval.ID,
) (*domainapproval.Approval, error) {
	var a domainapproval.Approval

	err := r.db.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			change_request_id,
			type,
			status,
			required,
			approver_id,
			decided_at,
			comment
		FROM approvals
		WHERE id = $1
		`,
		string(id),
	).Scan(
		&a.ID,
		&a.ChangeRequestID,
		&a.Type,
		&a.Status,
		&a.Required,
		&a.ApproverID,
		&a.DecidedAt,
		&a.Comment,
	)
	if err != nil {
		return nil, fmt.Errorf("get approval: %w", err)
	}

	return &a, nil
}

func (r *ApprovalRepository) ListByChangeRequestID(
	ctx context.Context,
	changeRequestID string,
) ([]domainapproval.Approval, error) {
	rows, err := r.db.pool.Query(
		ctx,
		`
		SELECT
			id,
			change_request_id,
			type,
			status,
			required,
			approver_id,
			decided_at,
			comment
		FROM approvals
		WHERE change_request_id = $1
		ORDER BY id
		`,
		changeRequestID,
	)
	if err != nil {
		return nil, fmt.Errorf("list approvals: %w", err)
	}
	defer rows.Close()

	approvals := make([]domainapproval.Approval, 0)
	for rows.Next() {
		var a domainapproval.Approval

		if err := rows.Scan(
			&a.ID,
			&a.ChangeRequestID,
			&a.Type,
			&a.Status,
			&a.Required,
			&a.ApproverID,
			&a.DecidedAt,
			&a.Comment,
		); err != nil {
			return nil, fmt.Errorf("scan approval: %w", err)
		}

		approvals = append(approvals, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approvals: %w", err)
	}

	return approvals, nil
}

func (r *ApprovalRepository) Save(
	ctx context.Context,
	a domainapproval.Approval,
) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("validate approval: %w", err)
	}

	result, err := r.db.pool.Exec(
		ctx,
		`
		UPDATE approvals
		SET
			status = $2,
			decided_at = $3,
			comment = $4
		WHERE id = $1
		`,
		string(a.ID),
		a.Status,
		a.DecidedAt,
		a.Comment,
	)
	if err != nil {
		return fmt.Errorf("save approval: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("approval %q not found", a.ID)
	}

	return nil
}

var _ ports.ApprovalRepository = (*ApprovalRepository)(nil)
