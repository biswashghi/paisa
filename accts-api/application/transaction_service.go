package application

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		insertedID, inserted, err := stores.Transactions.CreateIfNotExists(ctx, input)
		if err != nil {
			return err
		}
		if !inserted {
			existing, err := stores.Transactions.FindByExternalID(ctx, partner.ID, body.ExternalTransactionID)
			if err != nil {
				return err
			}
			if existing.PayloadHash != payloadHash {
				return domain.ConflictError("same external transaction id has a different payload")
			}
			eventID = existing.ID
			return nil
		}
		eventID = insertedID
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

func (s TransactionIngestionService) IngestManualTransaction(ctx context.Context, auth domain.AuthContext, body domain.ManualTransactionRequest) (domain.TransactionEvent, error) {
	externalID := strings.TrimSpace(body.ExternalTransactionID)
	if externalID == "" {
		externalID = fmt.Sprintf("manual-%d", time.Now().UnixNano())
	}
	customer := body.Customer
	if body.MemberID != "" {
		member, err := s.app.stores.Members.GetByID(ctx, auth.PartnerID, body.MemberID)
		if err != nil {
			return domain.TransactionEvent{}, err
		}
		customer.ExternalCustomerID = member.ExternalCustomerID
	}
	lines := body.LineItems
	if len(lines) == 0 {
		category := strings.TrimSpace(body.Category)
		if category == "" {
			category = "manual"
		}
		lines = []domain.LineItemInput{{
			ExternalLineID: "manual-line-1",
			Category:       category,
			Quantity:       1,
			SubtotalMinor:  body.SubtotalMinor,
			TaxMinor:       body.TaxMinor,
			TotalMinor:     body.TotalMinor,
			EligibleMinor:  body.EligibleMinor,
		}}
	}
	return s.IngestNormalizedTransaction(ctx, auth.PartnerID, domain.NormalizedTransaction{
		SourceSystem:          "manual",
		SourceLocationID:      body.LocationID,
		ExternalEventType:     "manual_purchase",
		IdempotencyKey:        externalID,
		ExternalTransactionID: externalID,
		Customer:              customer,
		MemberID:              body.MemberID,
		Type:                  domain.EventPurchase,
		Currency:              normalizeDefault(body.Currency, "USD"),
		SubtotalMinor:         body.SubtotalMinor,
		TaxMinor:              body.TaxMinor,
		TotalMinor:            body.TotalMinor,
		EligibleMinor:         body.EligibleMinor,
		OccurredAt:            body.OccurredAt,
		LineItems:             lines,
		RawPayload:            domain.JSONMap{"source": "cashier_manual"},
	})
}

func (s TransactionIngestionService) IngestNormalizedTransaction(ctx context.Context, partnerID string, body domain.NormalizedTransaction) (domain.TransactionEvent, error) {
	partner, err := s.app.stores.Partners.GetByID(ctx, partnerID)
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
	if body.EligibleMinor == 0 {
		body.EligibleMinor = body.SubtotalMinor
	}
	if body.TotalMinor == 0 {
		body.TotalMinor = body.SubtotalMinor + body.TaxMinor
	}
	if err := domain.ValidateTransactionType(body.Type); err != nil {
		return domain.TransactionEvent{}, err
	}
	memberID := body.MemberID
	if memberID == "" {
		resolved, err := MemberService{app: s.app}.ResolveOrCreateMemberForPartnerID(ctx, partner.ID, body.Customer)
		if err != nil {
			return domain.TransactionEvent{}, err
		}
		memberID = resolved.Member.ID
	}
	member, err := s.app.stores.Members.GetByID(ctx, partner.ID, memberID)
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	if err := domain.EnsureActiveStatus("member", member.Status); err != nil {
		return domain.TransactionEvent{}, err
	}

	legacyPayload := domain.TransactionIngestRequest{
		ExternalTransactionID:         body.ExternalTransactionID,
		ExternalCustomerID:            member.ExternalCustomerID,
		OriginalExternalTransactionID: body.OriginalExternalTransactionID,
		Type:                          body.Type,
		Currency:                      body.Currency,
		SubtotalMinor:                 body.SubtotalMinor,
		TaxMinor:                      body.TaxMinor,
		TotalMinor:                    body.TotalMinor,
		EligibleMinor:                 body.EligibleMinor,
		OccurredAt:                    body.OccurredAt,
		LineItems:                     body.LineItems,
	}
	payloadHash := domain.HashTransactionPayload(legacyPayload)
	occurredAt, err := domain.ParseOccurredAtStrict(body.OccurredAt)
	if err != nil {
		return domain.TransactionEvent{}, err
	}
	rawPayload := body.RawPayload
	if rawPayload == nil {
		rawPayload = domain.JSONMap{}
	}
	rawPayload["normalized"] = legacyPayload
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
		RawPayload:                    mustJSON(rawPayload),
		PayloadHash:                   payloadHash,
		SourceSystem:                  normalizeDefault(body.SourceSystem, "manual"),
		SourceConnectionID:            body.SourceConnectionID,
		SourceLocationID:              body.SourceLocationID,
		ExternalEventType:             body.ExternalEventType,
		IdempotencyKey:                body.IdempotencyKey,
	}

	var eventID string
	err = s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		insertedID, inserted, err := stores.Transactions.CreateIfNotExists(ctx, input)
		if err != nil {
			return err
		}
		if !inserted {
			existing, err := stores.Transactions.FindByExternalID(ctx, partner.ID, body.ExternalTransactionID)
			if err != nil {
				return err
			}
			if existing.PayloadHash != payloadHash {
				return domain.ConflictError("same external transaction id has a different payload")
			}
			eventID = existing.ID
			return nil
		}
		eventID = insertedID
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
