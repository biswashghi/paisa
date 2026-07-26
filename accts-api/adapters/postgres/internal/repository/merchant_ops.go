package repository

import (
	"context"
	"database/sql"
	"time"

	"accts-api/domain"
)

func (s LocationStore) Create(ctx context.Context, partnerID string, body domain.LocationRequest) (domain.PartnerLocation, error) {
	var location domain.PartnerLocation
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.partner_locations (partner_id, name, address, timezone, external_location_id)
		VALUES ($1, $2, NULLIF($3, ''), COALESCE(NULLIF($4, ''), 'America/Detroit'), NULLIF($5, ''))
		RETURNING id, partner_id, name, COALESCE(address, ''), timezone, status, COALESCE(external_location_id, ''), created_at, updated_at`,
		partnerID, body.Name, body.Address, body.Timezone, body.ExternalLocationID,
	).Scan(&location.ID, &location.PartnerID, &location.Name, &location.Address, &location.Timezone, &location.Status, &location.ExternalLocationID, &location.CreatedAt, &location.UpdatedAt)
	return location, AppErrorFromDB(err)
}

func (s LocationStore) List(ctx context.Context, partnerID string) ([]domain.PartnerLocation, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, name, COALESCE(address, ''), timezone, status, COALESCE(external_location_id, ''), created_at, updated_at
		FROM paisa.partner_locations
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	locations := []domain.PartnerLocation{}
	for rows.Next() {
		var location domain.PartnerLocation
		if err := rows.Scan(&location.ID, &location.PartnerID, &location.Name, &location.Address, &location.Timezone, &location.Status, &location.ExternalLocationID, &location.CreatedAt, &location.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		locations = append(locations, location)
	}
	return locations, AppErrorFromDB(rows.Err())
}

func (s CatalogStore) Create(ctx context.Context, partnerID string, body domain.CatalogItemRequest) (domain.CatalogItem, error) {
	var item domain.CatalogItem
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.catalog_items (
			partner_id, program_id, location_id, name, description, points_cost, reward_type, status, expires_after_minutes
		)
		VALUES ($1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, $4, NULLIF($5, ''), $6,
			COALESCE(NULLIF($7, ''), 'manual_discount'), COALESCE(NULLIF($8, ''), 'active'), COALESCE(NULLIF($9, 0), 15))
		RETURNING id, partner_id, COALESCE(program_id::text, ''), COALESCE(location_id::text, ''), name,
			COALESCE(description, ''), points_cost, reward_type, status, expires_after_minutes, created_at, updated_at`,
		partnerID, body.ProgramID, body.LocationID, body.Name, body.Description, body.PointsCost, body.RewardType, body.Status, body.ExpiresAfterMinutes,
	).Scan(&item.ID, &item.PartnerID, &item.ProgramID, &item.LocationID, &item.Name, &item.Description, &item.PointsCost, &item.RewardType, &item.Status, &item.ExpiresAfterMinutes, &item.CreatedAt, &item.UpdatedAt)
	return item, AppErrorFromDB(err)
}

func (s CatalogStore) List(ctx context.Context, partnerID string) ([]domain.CatalogItem, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, COALESCE(program_id::text, ''), COALESCE(location_id::text, ''), name,
			COALESCE(description, ''), points_cost, reward_type, status, expires_after_minutes, created_at, updated_at
		FROM paisa.catalog_items
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	return scanCatalogItems(rows)
}

func (s CatalogStore) Update(ctx context.Context, partnerID, itemID string, body domain.CatalogItemRequest) (domain.CatalogItem, error) {
	var item domain.CatalogItem
	err := s.q.QueryRowContext(ctx, `
		UPDATE paisa.catalog_items
		SET program_id = NULLIF($3, '')::uuid,
			location_id = NULLIF($4, '')::uuid,
			name = $5,
			description = NULLIF($6, ''),
			points_cost = $7,
			reward_type = COALESCE(NULLIF($8, ''), reward_type),
			status = COALESCE(NULLIF($9, ''), status),
			expires_after_minutes = COALESCE(NULLIF($10, 0), expires_after_minutes),
			updated_at = now()
		WHERE partner_id = $1 AND id = $2
		RETURNING id, partner_id, COALESCE(program_id::text, ''), COALESCE(location_id::text, ''), name,
			COALESCE(description, ''), points_cost, reward_type, status, expires_after_minutes, created_at, updated_at`,
		partnerID, itemID, body.ProgramID, body.LocationID, body.Name, body.Description, body.PointsCost, body.RewardType, body.Status, body.ExpiresAfterMinutes,
	).Scan(&item.ID, &item.PartnerID, &item.ProgramID, &item.LocationID, &item.Name, &item.Description, &item.PointsCost, &item.RewardType, &item.Status, &item.ExpiresAfterMinutes, &item.CreatedAt, &item.UpdatedAt)
	return item, AppErrorFromDB(err)
}

func (s CatalogStore) Get(ctx context.Context, partnerID, itemID string) (domain.CatalogItem, error) {
	var item domain.CatalogItem
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_id, COALESCE(program_id::text, ''), COALESCE(location_id::text, ''), name,
			COALESCE(description, ''), points_cost, reward_type, status, expires_after_minutes, created_at, updated_at
		FROM paisa.catalog_items
		WHERE partner_id = $1 AND id = $2`, partnerID, itemID,
	).Scan(&item.ID, &item.PartnerID, &item.ProgramID, &item.LocationID, &item.Name, &item.Description, &item.PointsCost, &item.RewardType, &item.Status, &item.ExpiresAfterMinutes, &item.CreatedAt, &item.UpdatedAt)
	return item, AppErrorFromDB(err)
}

func (s CatalogStore) AvailableForMember(ctx context.Context, partnerID, memberID string) ([]domain.CatalogItem, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT ci.id, ci.partner_id, COALESCE(ci.program_id::text, ''), COALESCE(ci.location_id::text, ''), ci.name,
			COALESCE(ci.description, ''), ci.points_cost, ci.reward_type, ci.status, ci.expires_after_minutes, ci.created_at, ci.updated_at
		FROM paisa.catalog_items ci
		LEFT JOIN paisa.program_enrollments pe ON pe.partner_id = ci.partner_id AND pe.member_id = $2 AND pe.status = 'active'
		WHERE ci.partner_id = $1 AND ci.status = 'active' AND (ci.program_id IS NULL OR ci.program_id = pe.program_id)
		ORDER BY ci.points_cost, ci.created_at`, partnerID, memberID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	return scanCatalogItems(rows)
}

func scanCatalogItems(rows *sql.Rows) ([]domain.CatalogItem, error) {
	items := []domain.CatalogItem{}
	for rows.Next() {
		var item domain.CatalogItem
		if err := rows.Scan(&item.ID, &item.PartnerID, &item.ProgramID, &item.LocationID, &item.Name, &item.Description, &item.PointsCost, &item.RewardType, &item.Status, &item.ExpiresAfterMinutes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		items = append(items, item)
	}
	return items, AppErrorFromDB(rows.Err())
}

func (s RedemptionStore) Create(ctx context.Context, partnerID, memberID, accountID, catalogItemID string, pointsCost int, code string, expiresAt time.Time) (domain.Redemption, error) {
	var redemption domain.Redemption
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.redemptions (
			partner_id, member_id, member_account_id, catalog_item_id, code, status, points_cost, reservation_expires_at
		)
		VALUES ($1, $2, $3, $4, $5, 'reserved', $6, $7)
		RETURNING id`, partnerID, memberID, accountID, catalogItemID, code, pointsCost, expiresAt,
	).Scan(&redemption.ID)
	if err != nil {
		return domain.Redemption{}, AppErrorFromDB(err)
	}
	return s.Get(ctx, partnerID, redemption.ID)
}

func (s RedemptionStore) Get(ctx context.Context, partnerID, redemptionID string) (domain.Redemption, error) {
	var redemption domain.Redemption
	var expires sql.NullTime
	err := s.q.QueryRowContext(ctx, `
		SELECT r.id, r.partner_id, r.member_id, r.member_account_id, r.catalog_item_id, ci.name, r.code, r.status,
			r.points_cost, r.reservation_expires_at, COALESCE(r.failure_reason, ''), r.created_at, r.updated_at
		FROM paisa.redemptions r
		JOIN paisa.catalog_items ci ON ci.id = r.catalog_item_id
		WHERE r.partner_id = $1 AND r.id = $2`, partnerID, redemptionID,
	).Scan(&redemption.ID, &redemption.PartnerID, &redemption.MemberID, &redemption.MemberAccountID, &redemption.CatalogItemID, &redemption.CatalogItemName, &redemption.Code, &redemption.Status, &redemption.PointsCost, &expires, &redemption.FailureReason, &redemption.CreatedAt, &redemption.UpdatedAt)
	if expires.Valid {
		redemption.ReservationExpiresAt = &expires.Time
	}
	return redemption, AppErrorFromDB(err)
}

func (s RedemptionStore) GetByCode(ctx context.Context, partnerID, code string) (domain.Redemption, error) {
	var id string
	err := s.q.QueryRowContext(ctx, `SELECT id FROM paisa.redemptions WHERE partner_id = $1 AND code = $2`, partnerID, code).Scan(&id)
	if err != nil {
		return domain.Redemption{}, AppErrorFromDB(err)
	}
	return s.Get(ctx, partnerID, id)
}

func (s RedemptionStore) UpdateStatus(ctx context.Context, partnerID, redemptionID, status, failureReason string) (domain.Redemption, error) {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.redemptions
		SET status = $3, failure_reason = COALESCE(NULLIF($4, ''), failure_reason), updated_at = now()
		WHERE partner_id = $1 AND id = $2`, partnerID, redemptionID, status, failureReason)
	if err != nil {
		return domain.Redemption{}, AppErrorFromDB(err)
	}
	return s.Get(ctx, partnerID, redemptionID)
}

func (s RedemptionStore) InsertEvent(ctx context.Context, redemptionID, status string, details domain.JSONMap) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO paisa.redemption_events (redemption_id, status, details)
		VALUES ($1, $2, $3)`, redemptionID, status, mustJSON(details))
	return AppErrorFromDB(err)
}

func (s RedemptionStore) List(ctx context.Context, partnerID string) ([]domain.Redemption, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT r.id
		FROM paisa.redemptions r
		WHERE r.partner_id = $1
		ORDER BY r.created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	redemptions := []domain.Redemption{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, AppErrorFromDB(err)
		}
		redemption, err := s.Get(ctx, partnerID, id)
		if err != nil {
			return nil, err
		}
		redemptions = append(redemptions, redemption)
	}
	return redemptions, AppErrorFromDB(rows.Err())
}

func (s RedemptionStore) Counts(ctx context.Context, partnerID string) (int, error) {
	var count int
	err := s.q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM paisa.redemptions
		WHERE partner_id = $1 AND status IN ('reserved', 'validated')`, partnerID).Scan(&count)
	return count, AppErrorFromDB(err)
}
