package repository

import (
	"context"

	"accts-api/domain"
)

func (s PartnerStore) Create(ctx context.Context, body domain.PartnerRequest) (domain.Partner, error) {
	var partner domain.Partner
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.partners (partner_key, name)
		VALUES ($1, $2)
		RETURNING id, partner_key, name, status, created_at, updated_at`,
		body.PartnerKey, body.Name,
	).Scan(&partner.ID, &partner.PartnerKey, &partner.Name, &partner.Status, &partner.CreatedAt, &partner.UpdatedAt)
	return partner, AppErrorFromDB(err)
}

func (s PartnerStore) List(ctx context.Context) ([]domain.Partner, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_key, name, status, created_at, updated_at
		FROM paisa.partners
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	partners := []domain.Partner{}
	for rows.Next() {
		var partner domain.Partner
		if err := rows.Scan(&partner.ID, &partner.PartnerKey, &partner.Name, &partner.Status, &partner.CreatedAt, &partner.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		partners = append(partners, partner)
	}
	return partners, AppErrorFromDB(rows.Err())
}

func (s PartnerStore) GetByKey(ctx context.Context, partnerKey string) (domain.Partner, error) {
	var partner domain.Partner
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_key, name, status, created_at, updated_at
		FROM paisa.partners
		WHERE partner_key = $1`, partnerKey,
	).Scan(&partner.ID, &partner.PartnerKey, &partner.Name, &partner.Status, &partner.CreatedAt, &partner.UpdatedAt)
	return partner, AppErrorFromDB(err)
}

func (s PartnerStore) GetByID(ctx context.Context, partnerID string) (domain.Partner, error) {
	var partner domain.Partner
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_key, name, status, created_at, updated_at
		FROM paisa.partners
		WHERE id = $1`, partnerID,
	).Scan(&partner.ID, &partner.PartnerKey, &partner.Name, &partner.Status, &partner.CreatedAt, &partner.UpdatedAt)
	return partner, AppErrorFromDB(err)
}
