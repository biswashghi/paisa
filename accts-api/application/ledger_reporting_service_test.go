package application

import (
	"context"
	"testing"

	"accts-api/domain"
	"accts-api/ports"
)

func TestLedgerReserveGuardPreventsNegativeAvailable(t *testing.T) {
	fake := newFakeAppStore()
	fake.balance.AvailablePoints = 10

	_, err := postLedgerEntry(context.Background(), fake.storeSet(), ports.LedgerEntryInput{
		PartnerID:       "partner_1",
		MemberAccountID: "account_1",
		EntryType:       domain.EntryRedemptionReserve,
		AvailableDelta:  -20,
		ReservedDelta:   20,
		SourceType:      "redemption",
		SourceID:        "00000000-0000-0000-0000-000000000001",
	})
	if !domain.IsErrorKind(err, domain.ErrorKindInvariant) {
		t.Fatalf("expected invariant error, got %v", err)
	}
}

func TestLedgerGetBalanceUsesDirectPartnerMemberLookup(t *testing.T) {
	fake := newFakeAppStore()
	fake.balance.AvailablePoints = 42

	balance, err := LedgerService{app: testApp(fake)}.GetBalance(context.Background(), "acme-demo", "member_1")
	if err != nil {
		t.Fatal(err)
	}
	if balance.AvailablePoints != 42 {
		t.Fatalf("expected balance from direct lookup, got %+v", balance)
	}
	if fake.balanceByMemberCalls != 1 {
		t.Fatalf("expected direct balance lookup, got %d calls", fake.balanceByMemberCalls)
	}
	if fake.accountIDCalls != 0 {
		t.Fatalf("expected no separate account lookup, got %d calls", fake.accountIDCalls)
	}
}

func TestGenerateLedgerLiabilityExportUsesReportingStore(t *testing.T) {
	fake := newFakeAppStore()
	fake.summary = domain.JSONMap{"liabilityPoints": 123}
	fake.exportID = "export_1"

	export, err := ReportingService{app: testApp(fake)}.GenerateLedgerLiabilityExport(context.Background(), domain.ExportRequest{
		PartnerKey:   "acme-demo",
		BusinessDate: "2026-07-25",
	})
	if err != nil {
		t.Fatal(err)
	}
	if export.ID != "export_1" || export.Summary["liabilityPoints"] != 123 {
		t.Fatalf("unexpected export: %+v", export)
	}
}
