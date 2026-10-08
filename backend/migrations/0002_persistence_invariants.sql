-- Enforce persistence invariants that cannot safely rely on application code alone.

-- A Change Request may contain a given controlled field only once.
CREATE UNIQUE INDEX change_request_changes_field_uq
    ON change_request_changes(change_request_id, field);

-- Approval decision state and timestamp must stay consistent.
ALTER TABLE approvals
    ADD CONSTRAINT approvals_decision_consistency_chk
    CHECK (
        (status = 'PENDING' AND decided_at IS NULL)
        OR
        (status IN ('APPROVED', 'REJECTED') AND decided_at IS NOT NULL)
    );
