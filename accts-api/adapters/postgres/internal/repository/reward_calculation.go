package repository

import (
	"context"
	"database/sql"

	"accts-api/domain"
	"accts-api/ports"
)

func (s RewardCalculationStore) Get(ctx context.Context, partnerID, transactionEventID string) (domain.RewardCalculation, error) {
	var calc domain.RewardCalculation
	var programID, ruleVersionID, failureReason sql.NullString
	var data []byte
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_id, transaction_event_id, program_id::text, rule_version_id::text, status,
			points_delta, COALESCE(basis_amount_minor, 0), calculation_data, failure_reason, created_at
		FROM paisa.reward_calculations
		WHERE partner_id = $1 AND transaction_event_id = $2`, partnerID, transactionEventID,
	).Scan(&calc.ID, &calc.PartnerID, &calc.TransactionEventID, &programID, &ruleVersionID, &calc.Status, &calc.PointsDelta, &calc.BasisAmountMinor, &data, &failureReason, &calc.CreatedAt)
	if err != nil {
		return calc, AppErrorFromDB(err)
	}
	if programID.Valid {
		calc.ProgramID = programID.String
	}
	if ruleVersionID.Valid {
		calc.RuleVersionID = ruleVersionID.String
	}
	if failureReason.Valid {
		calc.FailureReason = failureReason.String
	}
	calc.CalculationData = scanJSON(data)
	return calc, nil
}

func (s RewardCalculationStore) CreateSucceeded(ctx context.Context, input ports.RewardCalculationCreateInput) (string, error) {
	var calcID string
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.reward_calculations (
			partner_id, transaction_event_id, program_id, rule_version_id, status,
			points_delta, basis_amount_minor, calculation_data
		)
		VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, 'succeeded', $5, $6, $7)
		RETURNING id`,
		input.PartnerID, input.TransactionEventID, input.ProgramID, input.RuleVersionID,
		input.PointsDelta, input.BasisAmountMinor, mustJSON(input.CalculationData),
	).Scan(&calcID)
	return calcID, AppErrorFromDB(err)
}

func (s RewardCalculationStore) CreateFailed(ctx context.Context, event domain.RewardProcessingEvent, reason string) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO paisa.reward_calculations (partner_id, transaction_event_id, status, points_delta, basis_amount_minor, calculation_data, failure_reason)
		VALUES ($1, $2, 'failed', 0, $3, $4, $5)
		ON CONFLICT (transaction_event_id) DO NOTHING`, event.PartnerID, event.ID, event.EligibleMinor, mustJSON(domain.JSONMap{"failureReason": reason}), reason)
	return AppErrorFromDB(err)
}

func (s RewardCalculationStore) OriginalForRefund(ctx context.Context, event domain.RewardProcessingEvent) (domain.RewardProcessingEvent, ports.OriginalCalculation, error) {
	var originalID string
	if event.OriginalExternalTransactionID == "" {
		return domain.RewardProcessingEvent{}, ports.OriginalCalculation{}, AppErrorFromDB(sql.ErrNoRows)
	}
	err := s.q.QueryRowContext(ctx, `
		SELECT id
		FROM paisa.transaction_events
		WHERE partner_id = $1 AND external_transaction_id = $2`, event.PartnerID, event.OriginalExternalTransactionID,
	).Scan(&originalID)
	if err != nil {
		return domain.RewardProcessingEvent{}, ports.OriginalCalculation{}, AppErrorFromDB(err)
	}
	original, err := TransactionStore{q: s.q}.LoadForProcessing(ctx, originalID)
	if err != nil {
		return domain.RewardProcessingEvent{}, ports.OriginalCalculation{}, err
	}
	var calc ports.OriginalCalculation
	var programID, ruleVersionID sql.NullString
	var data []byte
	err = s.q.QueryRowContext(ctx, `
		SELECT id, program_id::text, rule_version_id::text, calculation_data
		FROM paisa.reward_calculations
		WHERE transaction_event_id = $1 AND status = 'succeeded'`, originalID,
	).Scan(&calc.ID, &programID, &ruleVersionID, &data)
	if err != nil {
		return domain.RewardProcessingEvent{}, ports.OriginalCalculation{}, AppErrorFromDB(err)
	}
	if programID.Valid {
		calc.ProgramID = programID.String
	}
	if ruleVersionID.Valid {
		calc.RuleVersionID = ruleVersionID.String
	}
	calc.CalculationData = scanJSON(data)
	return original, calc, nil
}

func (s RewardCalculationStore) PriorRefundedBasisForOriginal(ctx context.Context, event domain.RewardProcessingEvent) (int, error) {
	if event.OriginalExternalTransactionID == "" {
		return 0, nil
	}
	var total int
	err := s.q.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(COALESCE(rc.basis_amount_minor, 0)), 0)
		FROM paisa.transaction_events refund
		JOIN paisa.reward_calculations rc ON rc.transaction_event_id = refund.id
		WHERE refund.partner_id = $1
			AND refund.type = $2
			AND refund.original_external_transaction_id = $3
			AND refund.id <> $4
			AND rc.status = 'succeeded'`,
		event.PartnerID, domain.EventRefund, event.OriginalExternalTransactionID, event.ID,
	).Scan(&total)
	return total, AppErrorFromDB(err)
}
