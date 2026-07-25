package application

import (
	"context"

	"accts-api/domain"
	"accts-api/ports"
)

type TransactionIngestionService struct {
	app app
}

func (s TransactionIngestionService) IngestTransaction(ctx context.Context, partnerKey string, body domain.TransactionIngestRequest) (domain.TransactionEvent, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
		return domain.TransactionEvent{}, err
	}

	body.Type = normalizeDefault(body.Type, domain.EventPurchase)
	body.Currency = normalizeDefault(body.Currency, "USD")
	if body.ExternalTransactionID == "" {
		return domain.TransactionEvent{}, domain.InvalidError("externalTransactionId is required")
	}
	if err := domain.ValidateTransactionType(body.Type); err != nil {
		return domain.TransactionEvent{}, err
	}

	member, err := s.app.stores.Members.GetByExternalID(ctx, partner.ID, body.ExternalCustomerID)
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	if err := domain.EnsureActiveStatus("member", member.Status); err != nil {
		return domain.TransactionEvent{}, err
	}

	payloadHash := domain.HashTransactionPayload(body)
	existing, err := s.app.stores.Transactions.FindByExternalID(ctx, partner.ID, body.ExternalTransactionID)
	if err == nil {
		if existing.PayloadHash != payloadHash {
			return domain.TransactionEvent{}, domain.ConflictError("same external transaction id has a different payload")
		}
		return s.app.stores.Transactions.Get(ctx, partner.ID, existing.ID)
	}
	if !domain.IsErrorKind(err, domain.ErrorKindNotFound) {
		return domain.TransactionEvent{}, err
	}

	occurredAt, err := domain.ParseOccurredAtStrict(body.OccurredAt)
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	input := ports.TransactionCreateInput{
		PartnerID:                     partner.ID,
		MemberID:                      member.ID,
		ExternalTransactionID:         body.ExternalTransactionID,
		OriginalExternalTransactionID: body.OriginalExternalTransactionID,
		Type:                          body.Type,
		Currency:                      body.Currency,
		SubtotalMinor:                 body.SubtotalMinor,
		TaxMinor:                      body.TaxMinor,
		TotalMinor:                    body.TotalMinor,
		EligibleMinor:                 body.EligibleMinor,
		OccurredAt:                    occurredAt,
		RawPayload:                    mustJSON(domain.SanitizedTransactionPayload(body)),
		PayloadHash:                   payloadHash,
	}

	var eventID string
	err = s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		var err error
		eventID, err = stores.Transactions.Create(ctx, input)
		if err != nil {
			return err
		}
		lines := body.LineItems
		for i := range lines {
			if lines[i].Quantity == 0 {
				lines[i].Quantity = 1
			}
		}
		return stores.Transactions.InsertLineItems(ctx, eventID, lines)
	})
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	return s.app.stores.Transactions.Get(ctx, partner.ID, eventID)
}

func (s TransactionIngestionService) GetTransaction(ctx context.Context, partnerKey, eventID string) (domain.TransactionEvent, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	return s.app.stores.Transactions.Get(ctx, partner.ID, eventID)
}

func (s TransactionIngestionService) ListTransactions(ctx context.Context, partnerKey string) ([]domain.TransactionEvent, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	return s.app.stores.Transactions.List(ctx, partner.ID)
}

func (s TransactionIngestionService) GetCalculation(ctx context.Context, partnerKey, eventID string) (domain.RewardCalculation, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.RewardCalculation{}, err
	}
	return s.app.stores.RewardCalculations.Get(ctx, partner.ID, eventID)
}
