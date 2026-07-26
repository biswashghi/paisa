package repository

import (
	"context"

	"accts-api/domain"
	"accts-api/ports"
)

func (s LedgerStore) CreateBalanceSnapshot(ctx context.Context, accountID, partnerID string) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO paisa.balance_snapshots (member_account_id, partner_id)
		VALUES ($1, $2)`, accountID, partnerID)
	return AppErrorFromDB(err)
}

func (s LedgerStore) GetBalance(ctx context.Context, accountID string) (domain.BalanceSnapshot, error) {
	var balance domain.BalanceSnapshot
	err := s.q.QueryRowContext(ctx, `
		SELECT member_account_id, partner_id, available_points, reserved_points, expired_points, updated_at
		FROM paisa.balance_snapshots
		WHERE member_account_id = $1`, accountID,
	).Scan(&balance.MemberAccountID, &balance.PartnerID, &balance.AvailablePoints, &balance.ReservedPoints, &balance.ExpiredPoints, &balance.UpdatedAt)
	return balance, AppErrorFromDB(err)
}

func (s LedgerStore) GetBalanceByMember(ctx context.Context, partnerID, memberID string) (domain.BalanceSnapshot, error) {
	var balance domain.BalanceSnapshot
	err := s.q.QueryRowContext(ctx, `
		SELECT bs.member_account_id, bs.partner_id, bs.available_points, bs.reserved_points, bs.expired_points, bs.updated_at
		FROM paisa.balance_snapshots bs
		JOIN paisa.member_accounts ma ON ma.id = bs.member_account_id
		JOIN paisa.members m ON m.id = ma.member_id
		WHERE bs.partner_id = $1
			AND ma.member_id = $2
			AND ma.status = $3
			AND m.status = $3`,
		partnerID, memberID, domain.StatusActive,
	).Scan(&balance.MemberAccountID, &balance.PartnerID, &balance.AvailablePoints, &balance.ReservedPoints, &balance.ExpiredPoints, &balance.UpdatedAt)
	return balance, AppErrorFromDB(err)
}

func (s LedgerStore) LockBalance(ctx context.Context, accountID string) (domain.BalanceSnapshot, error) {
	var balance domain.BalanceSnapshot
	err := s.q.QueryRowContext(ctx, `
		SELECT member_account_id, partner_id, available_points, reserved_points, expired_points, updated_at
		FROM paisa.balance_snapshots
		WHERE member_account_id = $1
		FOR UPDATE`, accountID,
	).Scan(&balance.MemberAccountID, &balance.PartnerID, &balance.AvailablePoints, &balance.ReservedPoints, &balance.ExpiredPoints, &balance.UpdatedAt)
	return balance, AppErrorFromDB(err)
}

func (s LedgerStore) ListEntries(ctx context.Context, accountID string) ([]domain.LedgerEntry, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, member_account_id, COALESCE(program_id::text, ''), entry_type,
			available_delta, reserved_delta, expired_delta, source_type, source_id, COALESCE(reason, ''),
			created_by_type, COALESCE(created_by_id::text, ''), created_at
		FROM paisa.ledger_entries
		WHERE member_account_id = $1
		ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	entries := []domain.LedgerEntry{}
	for rows.Next() {
		var entry domain.LedgerEntry
		if err := rows.Scan(&entry.ID, &entry.PartnerID, &entry.MemberAccountID, &entry.ProgramID, &entry.EntryType, &entry.AvailableDelta, &entry.ReservedDelta, &entry.ExpiredDelta, &entry.SourceType, &entry.SourceID, &entry.Reason, &entry.CreatedByType, &entry.CreatedByID, &entry.CreatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		entries = append(entries, entry)
	}
	return entries, AppErrorFromDB(rows.Err())
}

func (s LedgerStore) InsertEntry(ctx context.Context, input ports.LedgerEntryInput) (string, error) {
	var entryID string
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.ledger_entries (
			partner_id, member_account_id, program_id, entry_type,
			available_delta, reserved_delta, expired_delta,
			source_type, source_id, reason, created_by_type, created_by_id
		)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, '')::uuid)
		RETURNING id`,
		input.PartnerID, input.MemberAccountID, input.ProgramID, input.EntryType,
		input.AvailableDelta, input.ReservedDelta, input.ExpiredDelta,
		input.SourceType, input.SourceID, input.Reason, normalizeDefault(input.CreatedByType, "system"), input.CreatedByID,
	).Scan(&entryID)
	return entryID, AppErrorFromDB(err)
}

func (s LedgerStore) UpdateBalance(ctx context.Context, balance domain.BalanceSnapshot) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.balance_snapshots
		SET available_points = $1, reserved_points = $2, expired_points = $3, updated_at = now()
		WHERE member_account_id = $4`, balance.AvailablePoints, balance.ReservedPoints, balance.ExpiredPoints, balance.MemberAccountID)
	return AppErrorFromDB(err)
}

func (s LedgerStore) PostLedgerEntry(ctx context.Context, input ports.LedgerEntryInput) (ports.LedgerPostResult, error) {
	balance, err := s.LockBalance(ctx, input.MemberAccountID)
	if err != nil {
		return ports.LedgerPostResult{}, err
	}
	next, err := domain.ApplyLedgerDelta(balance, input.EntryType, input.AvailableDelta, input.ReservedDelta, input.ExpiredDelta)
	if err != nil {
		return ports.LedgerPostResult{}, err
	}
	entryID, err := s.InsertEntry(ctx, input)
	if err != nil {
		return ports.LedgerPostResult{}, err
	}
	if err := s.UpdateBalance(ctx, next); err != nil {
		return ports.LedgerPostResult{}, err
	}
	return ports.LedgerPostResult{LedgerEntryID: entryID, Balance: next}, nil
}
