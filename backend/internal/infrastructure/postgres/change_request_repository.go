package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type ChangeRequestRepository struct {
	db *DB
}

func NewChangeRequestRepository(db *DB) *ChangeRequestRepository {
	return &ChangeRequestRepository{db: db}
}

func (r *ChangeRequestRepository) Create(
	ctx context.Context,
	cr domainchange.ChangeRequest,
) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin change request creation: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO change_requests (
			id,
			asset_id,
			base_version,
			status,
			initiator,
			title,
			description,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		cr.ID.String(),
		cr.AssetID,
		cr.BaseVersion,
		cr.Status,
		cr.Initiator,
		cr.Title,
		cr.Description,
		cr.CreatedAt,
		cr.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert change request: %w", err)
	}

	for _, change := range cr.Changes {
		oldValue, err := encodeFieldValue(change.OldValue)
		if err != nil {
			return fmt.Errorf("encode old field value: %w", err)
		}

		newValue, err := encodeFieldValue(change.NewValue)
		if err != nil {
			return fmt.Errorf("encode new field value: %w", err)
		}

		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO change_request_changes (
				change_request_id,
				id,
				field,
				old_value,
				new_value
			)
			VALUES ($1, $2, $3, $4, $5)
			`,
			cr.ID.String(),
			change.ID,
			change.Field,
			oldValue,
			newValue,
		)
		if err != nil {
			return fmt.Errorf("insert change %q: %w", change.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit change request creation: %w", err)
	}

	return nil
}

func (r *ChangeRequestRepository) GetByID(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	var cr domainchange.ChangeRequest

	err := r.db.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			asset_id,
			base_version,
			status,
			initiator,
			title,
			description,
			created_at,
			updated_at
		FROM change_requests
		WHERE id = $1
		`,
		id.String(),
	).Scan(
		&cr.ID,
		&cr.AssetID,
		&cr.BaseVersion,
		&cr.Status,
		&cr.Initiator,
		&cr.Title,
		&cr.Description,
		&cr.CreatedAt,
		&cr.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("change request %q: %w", id, ports.ErrNotFound)
		}

		return nil, fmt.Errorf("get change request: %w", err)
	}

	rows, err := r.db.pool.Query(
		ctx,
		`
		SELECT
			id,
			field,
			old_value,
			new_value
		FROM change_request_changes
		WHERE change_request_id = $1
		ORDER BY id
		`,
		id.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("get change request changes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			changeID string
			field    string
			oldRaw   []byte
			newRaw   []byte
		)

		if err := rows.Scan(
			&changeID,
			&field,
			&oldRaw,
			&newRaw,
		); err != nil {
			return nil, fmt.Errorf("scan change request change: %w", err)
		}

		fieldName, err := domainasset.ParseFieldName(field)
		if err != nil {
			return nil, fmt.Errorf("parse field name: %w", err)
		}

		oldValue, err := decodeFieldValue(oldRaw)
		if err != nil {
			return nil, fmt.Errorf("decode old value: %w", err)
		}

		newValue, err := decodeFieldValue(newRaw)
		if err != nil {
			return nil, fmt.Errorf("decode new value: %w", err)
		}

		if err := oldValue.ValidateFor(fieldName); err != nil {
			return nil, fmt.Errorf("validate old value: %w", err)
		}

		if err := newValue.ValidateFor(fieldName); err != nil {
			return nil, fmt.Errorf("validate new value: %w", err)
		}

		cr.Changes = append(cr.Changes, domainchange.FieldChange{
			ID:       changeID,
			Field:    fieldName,
			OldValue: oldValue,
			NewValue: newValue,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate change request changes: %w", err)
	}

	return &cr, nil
}

func (r *ChangeRequestRepository) Save(
	ctx context.Context,
	cr domainchange.ChangeRequest,
) error {
	commandTag, err := r.db.pool.Exec(
		ctx,
		`
		UPDATE change_requests
		SET
			status = $2,
			updated_at = $3
		WHERE id = $1
		`,
		cr.ID.String(),
		cr.Status,
		cr.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save change request: %w", err)
	}

	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("change request %q not found", cr.ID)
	}

	return nil
}

var _ ports.ChangeRequestRepository = (*ChangeRequestRepository)(nil)
