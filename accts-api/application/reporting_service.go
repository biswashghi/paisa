package application

import (
	"context"
	"time"

	"accts-api/domain"
)

type ReportingService struct {
	app app
}

func (s ReportingService) GenerateLedgerLiabilityExport(ctx context.Context, body domain.ExportRequest) (domain.LedgerExport, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, body.PartnerKey)
	if err != nil {
		return domain.LedgerExport{}, err
	}
	if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
		return domain.LedgerExport{}, err
	}
	businessDate := body.BusinessDate
	if businessDate == "" {
		businessDate = time.Now().Format("2006-01-02")
	}
	summary, err := s.app.stores.Reporting.LedgerSummary(ctx, partner.ID, businessDate)
	if err != nil {
		return domain.LedgerExport{}, err
	}
	exportID, err := s.app.stores.Reporting.UpsertLedgerLiabilityExport(ctx, partner.ID, businessDate, summary)
	if err != nil {
		return domain.LedgerExport{}, err
	}
	return domain.LedgerExport{
		ID:           exportID,
		PartnerID:    partner.ID,
		BusinessDate: businessDate,
		Status:       "complete",
		Summary:      summary,
	}, nil
}

func (s ReportingService) ListLedgerLiabilityExports(ctx context.Context, partnerKey string) ([]domain.LedgerExport, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	return s.app.stores.Reporting.ListLedgerLiabilityExports(ctx, partner.ID)
}
