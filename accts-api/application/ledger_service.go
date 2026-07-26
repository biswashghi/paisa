package application

import (
	"context"

	"accts-api/domain"
	"accts-api/ports"

	"github.com/google/uuid"
)

type LedgerService struct {
	app app
}

func (s LedgerService) GetBalance(ctx context.Context, partnerKey, memberID string) (domain.BalanceSnapshot, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.BalanceSnapshot{}, err
	}
	return s.app.stores.Ledger.GetBalanceByMember(ctx, partner.ID, memberID)
}

func (s LedgerService) GetLedger(ctx context.Context, partnerKey, memberID string) ([]domain.LedgerEntry, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	accountID, err := s.app.stores.Members.AccountID(ctx, partner.ID, memberID)
	if err != nil {
		return nil, err
	}
	return s.app.stores.Ledger.ListEntries(ctx, accountID)
}

func (s LedgerService) CreateAdjustment(ctx context.Context, partnerKey, memberID string, body domain.AdjustmentRequest) (domain.AdjustmentResult, error) {
	var result domain.AdjustmentResult
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}
		if body.Reason == "" {
			return domain.InvalidError("adjustment reason is required")
		}
		sourceID := body.SourceID
		if sourceID == "" {
			sourceID = uuid.NewString()
		}
		accountID, err := stores.Members.AccountID(ctx, partner.ID, memberID)
		if err != nil {
			return err
		}
		entryID, err := postLedgerEntry(ctx, stores, ports.LedgerEntryInput{
			PartnerID:       partner.ID,
			MemberAccountID: accountID,
			EntryType:       domain.EntryAdjustment,
			AvailableDelta:  body.AvailableDelta,
			ReservedDelta:   body.ReservedDelta,
			ExpiredDelta:    body.ExpiredDelta,
			SourceType:      "manual_adjustment",
			SourceID:        sourceID,
			Reason:          body.Reason,
			CreatedByType:   "system",
		})
		if err != nil {
			return err
		}
		result = domain.AdjustmentResult{LedgerEntryID: entryID}
		return nil
	})
	return result, err
}

func postLedgerEntry(ctx context.Context, stores ports.StoreSet, input ports.LedgerEntryInput) (string, error) {
	result, err := stores.Ledger.PostLedgerEntry(ctx, input)
	return result.LedgerEntryID, err
}
