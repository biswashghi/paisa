package repository

import (
	"context"
	"database/sql"

	"accts-api/domain"
)

func (s IntegrationStore) Create(ctx context.Context, partnerID string, body domain.IntegrationConnectionRequest) (domain.IntegrationConnection, error) {
	var connection domain.IntegrationConnection
	var lastSync sql.NullTime
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.integration_connections (
			partner_id, provider, status, external_merchant_id, external_location_id, metadata
		)
		VALUES ($1, $2, COALESCE(NULLIF($3, ''), 'pending'), NULLIF($4, ''), NULLIF($5, ''), COALESCE($6::jsonb, '{}'::jsonb))
		RETURNING id, partner_id, provider, status, COALESCE(external_merchant_id, ''), COALESCE(external_location_id, ''),
			metadata, last_sync_at, created_at, updated_at`,
		partnerID, body.Provider, body.Status, body.ExternalMerchantID, body.ExternalLocationID, mustJSON(body.Metadata),
	).Scan(&connection.ID, &connection.PartnerID, &connection.Provider, &connection.Status, &connection.ExternalMerchantID, &connection.ExternalLocationID, &connection.Metadata, &lastSync, &connection.CreatedAt, &connection.UpdatedAt)
	if lastSync.Valid {
		connection.LastSyncAt = &lastSync.Time
	}
	return connection, AppErrorFromDB(err)
}

func (s IntegrationStore) List(ctx context.Context, partnerID string) ([]domain.IntegrationConnection, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, provider, status, COALESCE(external_merchant_id, ''), COALESCE(external_location_id, ''),
			metadata, last_sync_at, created_at, updated_at
		FROM paisa.integration_connections
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	connections := []domain.IntegrationConnection{}
	for rows.Next() {
		var connection domain.IntegrationConnection
		var raw []byte
		var lastSync sql.NullTime
		if err := rows.Scan(&connection.ID, &connection.PartnerID, &connection.Provider, &connection.Status, &connection.ExternalMerchantID, &connection.ExternalLocationID, &raw, &lastSync, &connection.CreatedAt, &connection.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		connection.Metadata = scanJSON(raw)
		if lastSync.Valid {
			connection.LastSyncAt = &lastSync.Time
		}
		connections = append(connections, connection)
	}
	return connections, AppErrorFromDB(rows.Err())
}

func (s IntegrationStore) MarkSynced(ctx context.Context, partnerID, connectionID string) (domain.IntegrationConnection, error) {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.integration_connections
		SET status = 'connected', last_sync_at = now(), updated_at = now()
		WHERE partner_id = $1 AND id = $2`, partnerID, connectionID)
	if err != nil {
		return domain.IntegrationConnection{}, AppErrorFromDB(err)
	}
	connections, err := s.List(ctx, partnerID)
	if err != nil {
		return domain.IntegrationConnection{}, err
	}
	for _, connection := range connections {
		if connection.ID == connectionID {
			return connection, nil
		}
	}
	return domain.IntegrationConnection{}, domain.InvalidError("integration connection not found")
}

func (s IntegrationStore) WarningCount(ctx context.Context, partnerID string) (int, error) {
	var count int
	err := s.q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM paisa.integration_connections
		WHERE partner_id = $1 AND status IN ('pending', 'failed')`, partnerID).Scan(&count)
	return count, AppErrorFromDB(err)
}

func (s CampaignStore) Create(ctx context.Context, partnerID string, body domain.CampaignRequest) (domain.Campaign, error) {
	var campaign domain.Campaign
	var startsAt, endsAt sql.NullTime
	var raw []byte
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.campaigns (
			partner_id, name, description, status, starts_at, ends_at, required_visit_count, reward_catalog_item_id, metadata
		)
		VALUES ($1, $2, NULLIF($3, ''), COALESCE(NULLIF($4, ''), 'draft'), NULLIF($5, '')::timestamptz,
			NULLIF($6, '')::timestamptz, COALESCE(NULLIF($7, 0), 3), NULLIF($8, '')::uuid, COALESCE($9::jsonb, '{}'::jsonb))
		RETURNING id, partner_id, name, COALESCE(description, ''), status, starts_at, ends_at,
			required_visit_count, COALESCE(reward_catalog_item_id::text, ''), metadata, created_at, updated_at`,
		partnerID, body.Name, body.Description, body.Status, body.StartsAt, body.EndsAt, body.RequiredVisitCount, body.RewardCatalogItemID, mustJSON(body.Metadata),
	).Scan(&campaign.ID, &campaign.PartnerID, &campaign.Name, &campaign.Description, &campaign.Status, &startsAt, &endsAt, &campaign.RequiredVisitCount, &campaign.RewardCatalogItemID, &raw, &campaign.CreatedAt, &campaign.UpdatedAt)
	if startsAt.Valid {
		campaign.StartsAt = &startsAt.Time
	}
	if endsAt.Valid {
		campaign.EndsAt = &endsAt.Time
	}
	campaign.Metadata = scanJSON(raw)
	return campaign, AppErrorFromDB(err)
}

func (s CampaignStore) List(ctx context.Context, partnerID string) ([]domain.Campaign, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, name, COALESCE(description, ''), status, starts_at, ends_at,
			required_visit_count, COALESCE(reward_catalog_item_id::text, ''), metadata, created_at, updated_at
		FROM paisa.campaigns
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	campaigns := []domain.Campaign{}
	for rows.Next() {
		var campaign domain.Campaign
		var startsAt, endsAt sql.NullTime
		var raw []byte
		if err := rows.Scan(&campaign.ID, &campaign.PartnerID, &campaign.Name, &campaign.Description, &campaign.Status, &startsAt, &endsAt, &campaign.RequiredVisitCount, &campaign.RewardCatalogItemID, &raw, &campaign.CreatedAt, &campaign.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		if startsAt.Valid {
			campaign.StartsAt = &startsAt.Time
		}
		if endsAt.Valid {
			campaign.EndsAt = &endsAt.Time
		}
		campaign.Metadata = scanJSON(raw)
		campaigns = append(campaigns, campaign)
	}
	return campaigns, AppErrorFromDB(rows.Err())
}
