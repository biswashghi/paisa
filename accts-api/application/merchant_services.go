package application

import (
	"context"
	"strings"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

type LocationService struct{ app app }
type CatalogService struct{ app app }
type RedemptionService struct{ app app }
type IntegrationService struct{ app app }
type DashboardService struct{ app app }
type CampaignService struct{ app app }

func (s LocationService) CreateLocation(ctx context.Context, auth domain.AuthContext, body domain.LocationRequest) (domain.PartnerLocation, error) {
	if strings.TrimSpace(body.Name) == "" {
		return domain.PartnerLocation{}, domain.InvalidError("location name is required")
	}
	return s.app.stores.Locations.Create(ctx, auth.PartnerID, body)
}

func (s LocationService) ListLocations(ctx context.Context, auth domain.AuthContext) ([]domain.PartnerLocation, error) {
	return s.app.stores.Locations.List(ctx, auth.PartnerID)
}

func (s CatalogService) CreateCatalogItem(ctx context.Context, auth domain.AuthContext, body domain.CatalogItemRequest) (domain.CatalogItem, error) {
	if strings.TrimSpace(body.Name) == "" {
		return domain.CatalogItem{}, domain.InvalidError("catalog item name is required")
	}
	if body.PointsCost <= 0 {
		return domain.CatalogItem{}, domain.InvalidError("pointsCost must be greater than 0")
	}
	return s.app.stores.Catalog.Create(ctx, auth.PartnerID, body)
}

func (s CatalogService) ListCatalogItems(ctx context.Context, auth domain.AuthContext) ([]domain.CatalogItem, error) {
	return s.app.stores.Catalog.List(ctx, auth.PartnerID)
}

func (s CatalogService) UpdateCatalogItem(ctx context.Context, auth domain.AuthContext, itemID string, body domain.CatalogItemRequest) (domain.CatalogItem, error) {
	if strings.TrimSpace(body.Name) == "" {
		return domain.CatalogItem{}, domain.InvalidError("catalog item name is required")
	}
	if body.PointsCost <= 0 {
		return domain.CatalogItem{}, domain.InvalidError("pointsCost must be greater than 0")
	}
	return s.app.stores.Catalog.Update(ctx, auth.PartnerID, itemID, body)
}

func (s CatalogService) AvailableRewards(ctx context.Context, auth domain.AuthContext, memberID string) ([]domain.CatalogItem, error) {
	return s.app.stores.Catalog.AvailableForMember(ctx, auth.PartnerID, memberID)
}

func (s RedemptionService) CreateRedemption(ctx context.Context, auth domain.AuthContext, body domain.RedemptionRequest) (domain.RedemptionActionResult, error) {
	var result domain.RedemptionActionResult
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		if body.MemberID == "" || body.CatalogItemID == "" {
			return domain.InvalidError("memberId and catalogItemId are required")
		}
		item, err := stores.Catalog.Get(ctx, auth.PartnerID, body.CatalogItemID)
		if err != nil {
			return err
		}
		if item.Status != domain.StatusActive {
			return domain.InvariantError("catalog item is not active")
		}
		accountID, err := stores.Members.AccountID(ctx, auth.PartnerID, body.MemberID)
		if err != nil {
			return err
		}
		code := domain.GenerateToken("rdm")
		expiresAt := time.Now().UTC().Add(time.Duration(item.ExpiresAfterMinutes) * time.Minute)
		redemption, err := stores.Redemptions.Create(ctx, auth.PartnerID, body.MemberID, accountID, item.ID, item.PointsCost, code, expiresAt)
		if err != nil {
			return err
		}
		entryID, err := postLedgerEntry(ctx, stores, ports.LedgerEntryInput{
			PartnerID:       auth.PartnerID,
			MemberAccountID: accountID,
			ProgramID:       item.ProgramID,
			EntryType:       domain.EntryRedemptionReserve,
			AvailableDelta:  -item.PointsCost,
			ReservedDelta:   item.PointsCost,
			SourceType:      "redemption",
			SourceID:        redemption.ID,
			Reason:          "redemption reserve",
			CreatedByType:   auth.ActorType,
			CreatedByID:     auth.ActorID,
		})
		if err != nil {
			_, _ = stores.Redemptions.UpdateStatus(ctx, auth.PartnerID, redemption.ID, domain.RedemptionReleased, err.Error())
			return err
		}
		if err := stores.Redemptions.InsertEvent(ctx, redemption.ID, domain.RedemptionReserved, domain.JSONMap{"code": redemption.Code}); err != nil {
			return err
		}
		balance, err := stores.Ledger.GetBalance(ctx, accountID)
		if err != nil {
			return err
		}
		result = domain.RedemptionActionResult{Redemption: redemption, LedgerEntryID: entryID, Balance: balance}
		return nil
	})
	return result, err
}

func (s RedemptionService) ValidateRedemption(ctx context.Context, auth domain.AuthContext, redemptionID string) (domain.RedemptionActionResult, error) {
	return s.changeStatusOnly(ctx, auth, redemptionID, domain.RedemptionValidated)
}

func (s RedemptionService) CaptureRedemption(ctx context.Context, auth domain.AuthContext, redemptionID string) (domain.RedemptionActionResult, error) {
	var result domain.RedemptionActionResult
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		redemption, err := stores.Redemptions.Get(ctx, auth.PartnerID, redemptionID)
		if err != nil {
			return err
		}
		if redemption.Status != domain.RedemptionReserved && redemption.Status != domain.RedemptionValidated {
			return domain.InvariantError("only reserved or validated redemptions can be captured")
		}
		entryID, err := postLedgerEntry(ctx, stores, ports.LedgerEntryInput{
			PartnerID:       auth.PartnerID,
			MemberAccountID: redemption.MemberAccountID,
			EntryType:       domain.EntryRedemptionCapture,
			ReservedDelta:   -redemption.PointsCost,
			SourceType:      "redemption",
			SourceID:        redemption.ID,
			Reason:          "redemption capture",
			CreatedByType:   auth.ActorType,
			CreatedByID:     auth.ActorID,
		})
		if err != nil {
			return err
		}
		redemption, err = stores.Redemptions.UpdateStatus(ctx, auth.PartnerID, redemption.ID, domain.RedemptionCaptured, "")
		if err != nil {
			return err
		}
		if err := stores.Redemptions.InsertEvent(ctx, redemption.ID, domain.RedemptionCaptured, domain.JSONMap{}); err != nil {
			return err
		}
		balance, err := stores.Ledger.GetBalance(ctx, redemption.MemberAccountID)
		if err != nil {
			return err
		}
		result = domain.RedemptionActionResult{Redemption: redemption, LedgerEntryID: entryID, Balance: balance}
		return nil
	})
	return result, err
}

func (s RedemptionService) ReleaseRedemption(ctx context.Context, auth domain.AuthContext, redemptionID string) (domain.RedemptionActionResult, error) {
	var result domain.RedemptionActionResult
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		redemption, err := stores.Redemptions.Get(ctx, auth.PartnerID, redemptionID)
		if err != nil {
			return err
		}
		if redemption.Status != domain.RedemptionReserved && redemption.Status != domain.RedemptionValidated {
			return domain.InvariantError("only reserved or validated redemptions can be released")
		}
		entryID, err := postLedgerEntry(ctx, stores, ports.LedgerEntryInput{
			PartnerID:       auth.PartnerID,
			MemberAccountID: redemption.MemberAccountID,
			EntryType:       domain.EntryReservationRelease,
			AvailableDelta:  redemption.PointsCost,
			ReservedDelta:   -redemption.PointsCost,
			SourceType:      "redemption",
			SourceID:        redemption.ID,
			Reason:          "redemption release",
			CreatedByType:   auth.ActorType,
			CreatedByID:     auth.ActorID,
		})
		if err != nil {
			return err
		}
		redemption, err = stores.Redemptions.UpdateStatus(ctx, auth.PartnerID, redemption.ID, domain.RedemptionReleased, "")
		if err != nil {
			return err
		}
		if err := stores.Redemptions.InsertEvent(ctx, redemption.ID, domain.RedemptionReleased, domain.JSONMap{}); err != nil {
			return err
		}
		balance, err := stores.Ledger.GetBalance(ctx, redemption.MemberAccountID)
		if err != nil {
			return err
		}
		result = domain.RedemptionActionResult{Redemption: redemption, LedgerEntryID: entryID, Balance: balance}
		return nil
	})
	return result, err
}

func (s RedemptionService) changeStatusOnly(ctx context.Context, auth domain.AuthContext, redemptionID, status string) (domain.RedemptionActionResult, error) {
	var result domain.RedemptionActionResult
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		redemption, err := stores.Redemptions.Get(ctx, auth.PartnerID, redemptionID)
		if err != nil {
			return err
		}
		if redemption.Status != domain.RedemptionReserved {
			return domain.InvariantError("only reserved redemptions can be validated")
		}
		if redemption.ReservationExpiresAt != nil && redemption.ReservationExpiresAt.Before(time.Now().UTC()) {
			return domain.InvariantError("redemption reservation has expired")
		}
		redemption, err = stores.Redemptions.UpdateStatus(ctx, auth.PartnerID, redemption.ID, status, "")
		if err != nil {
			return err
		}
		if err := stores.Redemptions.InsertEvent(ctx, redemption.ID, status, domain.JSONMap{}); err != nil {
			return err
		}
		balance, err := stores.Ledger.GetBalance(ctx, redemption.MemberAccountID)
		if err != nil {
			return err
		}
		result = domain.RedemptionActionResult{Redemption: redemption, Balance: balance}
		return nil
	})
	return result, err
}

func (s RedemptionService) ListRedemptions(ctx context.Context, auth domain.AuthContext) ([]domain.Redemption, error) {
	return s.app.stores.Redemptions.List(ctx, auth.PartnerID)
}

func (s IntegrationService) ListConnections(ctx context.Context, auth domain.AuthContext) ([]domain.IntegrationConnection, error) {
	return s.app.stores.Integrations.List(ctx, auth.PartnerID)
}

func (s IntegrationService) StartSquareOAuth(ctx context.Context, auth domain.AuthContext) (domain.IntegrationConnection, error) {
	return s.app.stores.Integrations.Create(ctx, auth.PartnerID, domain.IntegrationConnectionRequest{
		Provider: "square",
		Status:   "pending",
		Metadata: domain.JSONMap{"mode": "oauth_start", "note": "Square OAuth credentials not configured in local v1"},
	})
}

func (s IntegrationService) CompleteSquareOAuth(ctx context.Context, auth domain.AuthContext, code string) (domain.IntegrationConnection, error) {
	return s.app.stores.Integrations.Create(ctx, auth.PartnerID, domain.IntegrationConnectionRequest{
		Provider:           "square",
		Status:             "connected",
		ExternalMerchantID: "square-merchant-" + strings.TrimSpace(code),
		Metadata:           domain.JSONMap{"mode": "oauth_callback", "importOnly": true},
	})
}

func (s IntegrationService) SyncConnection(ctx context.Context, auth domain.AuthContext, connectionID string) (domain.IntegrationConnection, error) {
	return s.app.stores.Integrations.MarkSynced(ctx, auth.PartnerID, connectionID)
}

func (s DashboardService) Summary(ctx context.Context, auth domain.AuthContext) (domain.DashboardSummary, error) {
	partner, err := s.app.stores.Partners.GetByID(ctx, auth.PartnerID)
	if err != nil {
		return domain.DashboardSummary{}, err
	}
	locations, _ := s.app.stores.Locations.List(ctx, auth.PartnerID)
	items, _ := s.app.stores.Catalog.List(ctx, auth.PartnerID)
	openRedemptions, _ := s.app.stores.Redemptions.Counts(ctx, auth.PartnerID)
	warnings, _ := s.app.stores.Integrations.WarningCount(ctx, auth.PartnerID)
	return domain.DashboardSummary{
		Partner:             partner,
		ActiveLocations:     len(locations),
		ActiveCatalogItems:  len(items),
		OpenRedemptions:     openRedemptions,
		IntegrationWarnings: warnings,
	}, nil
}

func (s CampaignService) CreateCampaign(ctx context.Context, auth domain.AuthContext, body domain.CampaignRequest) (domain.Campaign, error) {
	if strings.TrimSpace(body.Name) == "" {
		return domain.Campaign{}, domain.InvalidError("campaign name is required")
	}
	return s.app.stores.Campaigns.Create(ctx, auth.PartnerID, body)
}

func (s CampaignService) ListCampaigns(ctx context.Context, auth domain.AuthContext) ([]domain.Campaign, error) {
	return s.app.stores.Campaigns.List(ctx, auth.PartnerID)
}
