package repository

import (
	"context"

	"accts-api/domain"
)

func (s ReportingStore) LedgerSummary(ctx context.Context, partnerID, businessDate string) (domain.JSONMap, error) {
	var available, reserved, expired int
	err := s.q.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(available_delta), 0), COALESCE(SUM(reserved_delta), 0), COALESCE(SUM(expired_delta), 0)
		FROM paisa.ledger_entries
		WHERE partner_id = $1 AND created_at < ($2::date + interval '1 day')`, partnerID, businessDate,
	).Scan(&available, &reserved, &expired)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	var earned, refunded, adjusted int
	err = s.q.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(available_delta) FILTER (WHERE entry_type = 'earn'), 0),
			COALESCE(SUM(available_delta) FILTER (WHERE entry_type = 'refund'), 0),
			COALESCE(SUM(available_delta) FILTER (WHERE entry_type = 'adjustment'), 0)
		FROM paisa.ledger_entries
		WHERE partner_id = $1 AND created_at < ($2::date + interval '1 day')`, partnerID, businessDate,
	).Scan(&earned, &refunded, &adjusted)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	return domain.JSONMap{
		"availablePoints": available,
		"reservedPoints":  reserved,
		"expiredPoints":   expired,
		"earnedPoints":    earned,
		"refundedPoints":  refunded,
		"adjustedPoints":  adjusted,
		"liabilityPoints": available + reserved,
	}, nil
}

func (s ReportingStore) UpsertLedgerLiabilityExport(ctx context.Context, partnerID, businessDate string, summary domain.JSONMap) (string, error) {
	var exportID string
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.ledger_exports (partner_id, business_date, summary)
		VALUES ($1, $2, $3)
		ON CONFLICT (partner_id, business_date)
		DO UPDATE SET summary = EXCLUDED.summary, updated_at = now(), status = 'complete'
		RETURNING id`, partnerID, businessDate, mustJSON(summary),
	).Scan(&exportID)
	return exportID, AppErrorFromDB(err)
}

func (s ReportingStore) ListLedgerLiabilityExports(ctx context.Context, partnerID string) ([]domain.LedgerExport, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, business_date::text, status, COALESCE(file_path, ''), summary, created_at, updated_at
		FROM paisa.ledger_exports
		WHERE partner_id = $1
		ORDER BY business_date DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	exports := []domain.LedgerExport{}
	for rows.Next() {
		var export domain.LedgerExport
		var summaryBytes []byte
		if err := rows.Scan(&export.ID, &export.BusinessDate, &export.Status, &export.FilePath, &summaryBytes, &export.CreatedAt, &export.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		export.PartnerID = partnerID
		export.Summary = scanJSON(summaryBytes)
		exports = append(exports, export)
	}
	return exports, AppErrorFromDB(rows.Err())
}
