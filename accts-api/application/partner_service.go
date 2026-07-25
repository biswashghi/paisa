package application

import (
	"context"

	"accts-api/domain"
)

type PartnerService struct {
	app app
}

func (s PartnerService) Create(ctx context.Context, body domain.PartnerRequest) (domain.Partner, error) {
	return s.app.stores.Partners.Create(ctx, body)
}

func (s PartnerService) List(ctx context.Context) ([]domain.Partner, error) {
	return s.app.stores.Partners.List(ctx)
}

func (s PartnerService) GetByKey(ctx context.Context, partnerKey string) (domain.Partner, error) {
	return s.app.stores.Partners.GetByKey(ctx, partnerKey)
}
