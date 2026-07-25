package repository

import (
	"context"
	"database/sql"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

func (s TransactionStore) FindByExternalID(ctx context.Context, partnerID, externalTransactionID string) (ports.TransactionIdentity, error) {
	var identity ports.TransactionIdentity
	err := s.q.QueryRowContext(ctx, `
		SELECT id, payload_hash
		FROM paisa.transaction_events
		WHERE partner_id = $1 AND external_transaction_id = $2`,
		partnerID, externalTransactionID,
	).Scan(&identity.ID, &identity.PayloadHash)
	return identity, AppErrorFromDB(err)
}

func (s TransactionStore) Create(ctx context.Context, input ports.TransactionCreateInput) (string, error) {
	var eventID string
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.transaction_events (
			partner_id, member_id, external_transaction_id, original_external_transaction_id,
			type, currency, subtotal_minor, tax_minor, total_minor, eligible_minor,
			occurred_at, raw_payload, payload_hash
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`,
		input.PartnerID, input.MemberID, input.ExternalTransactionID, input.OriginalExternalTransactionID,
		input.Type, input.Currency, input.SubtotalMinor, input.TaxMinor, input.TotalMinor, input.EligibleMinor,
		input.OccurredAt, input.RawPayload, input.PayloadHash,
	).Scan(&eventID)
	return eventID, AppErrorFromDB(err)
}

func (s TransactionStore) InsertLineItems(ctx context.Context, eventID string, lines []domain.LineItemInput) error {
	for _, line := range lines {
		if _, err := s.q.ExecContext(ctx, `
			INSERT INTO paisa.transaction_line_items (
				transaction_event_id, external_line_id, sku, category, quantity,
				subtotal_minor, tax_minor, total_minor, eligible_minor
			)
			VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7, $8, $9)`,
			eventID, line.ExternalLineID, line.SKU, line.Category, line.Quantity,
			line.SubtotalMinor, line.TaxMinor, line.TotalMinor, line.EligibleMinor,
		); err != nil {
			return AppErrorFromDB(err)
		}
	}
	return nil
}

func (s TransactionStore) Get(ctx context.Context, partnerID, eventID string) (domain.TransactionEvent, error) {
	var event domain.TransactionEvent
	var rawPayload []byte
	var original sql.NullString
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_id, member_id, external_transaction_id, original_external_transaction_id,
			type, status, currency, COALESCE(subtotal_minor, 0), COALESCE(tax_minor, 0),
			COALESCE(total_minor, 0), COALESCE(eligible_minor, 0), occurred_at, raw_payload, created_at, updated_at
		FROM paisa.transaction_events
		WHERE id = $1 AND partner_id = $2`, eventID, partnerID,
	).Scan(&event.ID, &event.PartnerID, &event.MemberID, &event.ExternalTransactionID, &original,
		&event.Type, &event.Status, &event.Currency, &event.SubtotalMinor, &event.TaxMinor,
		&event.TotalMinor, &event.EligibleMinor, &event.OccurredAt, &rawPayload, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return event, AppErrorFromDB(err)
	}
	if original.Valid {
		event.OriginalExternalTransactionID = original.String
	}
	event.RawPayload = scanJSON(rawPayload)
	event.LineItems, err = s.lineItems(ctx, event.ID)
	return event, AppErrorFromDB(err)
}

func (s TransactionStore) List(ctx context.Context, partnerID string) ([]domain.TransactionEvent, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id
		FROM paisa.transaction_events
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	events := []domain.TransactionEvent{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, AppErrorFromDB(err)
		}
		event, err := s.Get(ctx, partnerID, id)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, AppErrorFromDB(rows.Err())
}

func (s TransactionStore) AcceptedIDs(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.q.QueryContext(ctx, `
		WITH candidates AS (
			SELECT id
			FROM paisa.transaction_events
			WHERE status = $1
			ORDER BY created_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE paisa.transaction_events te
		SET status = $3, updated_at = now()
		FROM candidates
		WHERE te.id = candidates.id
		RETURNING te.id`, domain.StatusAccepted, limit, domain.StatusProcessing)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	eventIDs := []string{}
	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return nil, AppErrorFromDB(err)
		}
		eventIDs = append(eventIDs, eventID)
	}
	return eventIDs, AppErrorFromDB(rows.Err())
}

func (s TransactionStore) ClaimAccepted(ctx context.Context, eventID string) error {
	result, err := s.q.ExecContext(ctx, `UPDATE paisa.transaction_events SET status = $1, updated_at = now() WHERE id = $2 AND status IN ($1, $3)`, domain.StatusProcessing, eventID, domain.StatusAccepted)
	if err != nil {
		return AppErrorFromDB(err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.InvariantError("event not accepted")
	}
	return nil
}

func (s TransactionStore) MarkProcessed(ctx context.Context, eventID string) error {
	_, err := s.q.ExecContext(ctx, `UPDATE paisa.transaction_events SET status = $1, updated_at = now() WHERE id = $2`, domain.StatusProcessed, eventID)
	return AppErrorFromDB(err)
}

func (s TransactionStore) MarkFailed(ctx context.Context, eventID string) error {
	_, err := s.q.ExecContext(ctx, `UPDATE paisa.transaction_events SET status = $1, updated_at = now() WHERE id = $2`, domain.StatusFailed, eventID)
	return AppErrorFromDB(err)
}

func (s TransactionStore) lineItems(ctx context.Context, eventID string) ([]domain.LineItemInput, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT COALESCE(external_line_id, ''), COALESCE(sku, ''), COALESCE(category, ''), quantity,
			COALESCE(subtotal_minor, 0), COALESCE(tax_minor, 0), COALESCE(total_minor, 0), COALESCE(eligible_minor, 0)
		FROM paisa.transaction_line_items
		WHERE transaction_event_id = $1
		ORDER BY created_at`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lines := []domain.LineItemInput{}
	for rows.Next() {
		var line domain.LineItemInput
		if err := rows.Scan(&line.ExternalLineID, &line.SKU, &line.Category, &line.Quantity, &line.SubtotalMinor, &line.TaxMinor, &line.TotalMinor, &line.EligibleMinor); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, AppErrorFromDB(rows.Err())
}

func (s TransactionStore) LoadForProcessing(ctx context.Context, eventID string) (domain.RewardProcessingEvent, error) {
	var event domain.RewardProcessingEvent
	var original sql.NullString
	var rawPayload []byte
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_id, member_id, external_transaction_id, original_external_transaction_id,
			type, currency, COALESCE(subtotal_minor, 0), COALESCE(tax_minor, 0), COALESCE(total_minor, 0),
			COALESCE(eligible_minor, 0), occurred_at, raw_payload
		FROM paisa.transaction_events
		WHERE id = $1`, eventID,
	).Scan(&event.ID, &event.PartnerID, &event.MemberID, &event.ExternalTransactionID, &original, &event.Type, &event.Currency,
		&event.SubtotalMinor, &event.TaxMinor, &event.TotalMinor, &event.EligibleMinor, &event.OccurredAt, &rawPayload)
	if err != nil {
		return event, AppErrorFromDB(err)
	}
	if original.Valid {
		event.OriginalExternalTransactionID = original.String
	}
	event.RawPayload = scanJSON(rawPayload)
	rows, err := s.q.QueryContext(ctx, `
		SELECT COALESCE(external_line_id, ''), COALESCE(sku, ''), COALESCE(category, ''), quantity,
			COALESCE(subtotal_minor, 0), COALESCE(tax_minor, 0), COALESCE(total_minor, 0), COALESCE(eligible_minor, 0)
		FROM paisa.transaction_line_items
		WHERE transaction_event_id = $1`, eventID)
	if err != nil {
		return event, AppErrorFromDB(err)
	}
	defer rows.Close()
	for rows.Next() {
		var line domain.LineItemInput
		if err := rows.Scan(&line.ExternalLineID, &line.SKU, &line.Category, &line.Quantity, &line.SubtotalMinor, &line.TaxMinor, &line.TotalMinor, &line.EligibleMinor); err != nil {
			return event, AppErrorFromDB(err)
		}
		event.Lines = append(event.Lines, line)
	}
	return event, AppErrorFromDB(rows.Err())
}

func (s TransactionStore) PriorProcessedPurchaseCount(ctx context.Context, partnerID, memberID, excludeEventID string) (int, error) {
	var count int
	err := s.q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM paisa.transaction_events
		WHERE partner_id = $1 AND member_id = $2 AND type = $3 AND status = $4 AND id <> $5`,
		partnerID, memberID, domain.EventPurchase, domain.StatusProcessed, excludeEventID,
	).Scan(&count)
	return count, AppErrorFromDB(err)
}

func (s TransactionStore) PriorProcessedPurchaseEligibleMinorSum(ctx context.Context, partnerID, memberID, excludeEventID string, since time.Time) (int, error) {
	var total int
	err := s.q.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(COALESCE(eligible_minor, 0)), 0)
		FROM paisa.transaction_events
		WHERE partner_id = $1 AND member_id = $2 AND type = $3 AND status = $4 AND id <> $5 AND occurred_at >= $6`,
		partnerID, memberID, domain.EventPurchase, domain.StatusProcessed, excludeEventID, since,
	).Scan(&total)
	return total, AppErrorFromDB(err)
}
