package application

import (
	"context"
	"testing"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

func TestProcessPurchaseCalculatesRewardsAndPostsLedger(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.ruleGraph = domain.RuleGraph{
		Groups: []domain.RuleGraphGroup{{ID: "group_1", Strategy: "stack"}},
		Rules: []domain.RuleGraphRule{{
			ID:       "rule_1",
			GroupID:  "group_1",
			RuleType: "points_per_dollar",
			Formula:  domain.JSONMap{"pointsPerDollar": 2},
		}},
	}

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if len(fake.createdCalculations) != 1 || fake.createdCalculations[0].PointsDelta != 200 {
		t.Fatalf("expected 200 point calculation, got %+v", fake.createdCalculations)
	}
	if len(fake.ledgerEntries) != 1 || fake.ledgerEntries[0].EntryType != domain.EntryEarn || fake.ledgerEntries[0].AvailableDelta != 200 {
		t.Fatalf("unexpected ledger entries: %+v", fake.ledgerEntries)
	}
	if fake.balance.AvailablePoints != 200 {
		t.Fatalf("expected available balance 200, got %d", fake.balance.AvailablePoints)
	}
}

func TestProcessPurchaseCalculatesDecimalPointsPerDollar(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.ruleGraph = domain.RuleGraph{
		Groups: []domain.RuleGraphGroup{{ID: "group_1", Strategy: domain.RuleStrategyStack}},
		Rules: []domain.RuleGraphRule{{
			ID:       "rule_1",
			GroupID:  "group_1",
			RuleType: domain.RuleTypePointsPerDollar,
			Formula:  domain.JSONMap{"pointsPerDollar": 0.01},
		}},
	}

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if got := fake.createdCalculations[0].PointsDelta; got != 1 {
		t.Fatalf("expected $100 at 0.01 points per dollar to earn 1 point, got %d", got)
	}
}

func TestProcessPurchaseStacksAssignedMemberAddOnRules(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.addOnRuleVersionIDs = []string{"addon_1"}
	fake.ruleGraphs = map[string]domain.RuleGraph{
		"version_1": {
			Groups: []domain.RuleGraphGroup{{ID: "base_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "base_rule", GroupID: "base_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 1}}},
		},
		"addon_1": {
			Groups: []domain.RuleGraphGroup{{ID: "addon_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "addon_rule", GroupID: "addon_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 2}}},
		},
	}

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if got := fake.createdCalculations[0].PointsDelta; got != 300 {
		t.Fatalf("expected base 100 + add-on 200, got %d", got)
	}
	data := fake.createdCalculations[0].CalculationData
	ids, ok := data["ruleVersionIds"].([]string)
	if !ok || len(ids) != 2 || ids[0] != "version_1" || ids[1] != "addon_1" {
		t.Fatalf("expected calculation trace to include base and add-on rule versions, got %#v", data["ruleVersionIds"])
	}
}

func TestProcessPurchaseWithoutAssignmentUsesBaseRulesOnly(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.ruleGraphs = map[string]domain.RuleGraph{
		"version_1": {
			Groups: []domain.RuleGraphGroup{{ID: "base_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "base_rule", GroupID: "base_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 1}}},
		},
	}

	if _, err := (RewardProcessingService{app: testApp(fake)}).ProcessTransactionEvents(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := fake.createdCalculations[0].PointsDelta; got != 100 {
		t.Fatalf("expected only base 100 points, got %d", got)
	}
}

func TestProcessRefundReversesOriginalCalculation(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_refund"}
	fake.processingEvents["evt_refund"] = domain.RewardProcessingEvent{
		ID:                            "evt_refund",
		PartnerID:                     "partner_1",
		MemberID:                      "member_1",
		OriginalExternalTransactionID: "txn_purchase",
		Type:                          domain.EventRefund,
		EligibleMinor:                 5000,
		OccurredAt:                    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.originalEvent = domain.RewardProcessingEvent{ID: "evt_purchase", EligibleMinor: 10000}
	fake.originalCalculation = ports.OriginalCalculation{
		ProgramID:     "program_1",
		RuleVersionID: "version_1",
		CalculationData: domain.JSONMap{"selectedAwards": []interface{}{map[string]interface{}{
			"ruleID":           "rule_1",
			"ruleVersionID":    "version_1",
			"basisAmountMinor": float64(10000),
			"points":           float64(100),
		}}},
	}
	fake.originalErr = nil
	fake.balance.AvailablePoints = 100

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if len(fake.createdCalculations) != 1 || fake.createdCalculations[0].PointsDelta != -50 {
		t.Fatalf("expected -50 point refund calculation, got %+v", fake.createdCalculations)
	}
	if fake.balance.AvailablePoints != 50 {
		t.Fatalf("expected available balance 50, got %d", fake.balance.AvailablePoints)
	}
}

func TestProcessRefundCapsCumulativeRefundBasis(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_refund"}
	fake.processingEvents["evt_refund"] = domain.RewardProcessingEvent{
		ID:                            "evt_refund",
		PartnerID:                     "partner_1",
		MemberID:                      "member_1",
		OriginalExternalTransactionID: "txn_purchase",
		Type:                          domain.EventRefund,
		EligibleMinor:                 5000,
		OccurredAt:                    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.originalEvent = domain.RewardProcessingEvent{ID: "evt_purchase", EligibleMinor: 10000}
	fake.originalCalculation = ports.OriginalCalculation{
		ProgramID:     "program_1",
		RuleVersionID: "version_1",
		CalculationData: domain.JSONMap{"selectedAwards": []interface{}{map[string]interface{}{
			"ruleID":           "rule_1",
			"ruleVersionID":    "version_1",
			"basisAmountMinor": float64(10000),
			"points":           float64(100),
		}}},
	}
	fake.originalErr = nil
	fake.priorRefundedBasis = 8000
	fake.balance.AvailablePoints = 100

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if len(fake.createdCalculations) != 1 || fake.createdCalculations[0].PointsDelta != -20 || fake.createdCalculations[0].BasisAmountMinor != 2000 {
		t.Fatalf("expected capped -20 point refund calculation, got %+v", fake.createdCalculations)
	}
	if fake.balance.AvailablePoints != 80 {
		t.Fatalf("expected available balance 80, got %d", fake.balance.AvailablePoints)
	}
}

func TestProcessPurchaseUsesPublishedRuleGraphCache(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase_1", "evt_purchase_2"}
	for _, id := range fake.acceptedIDs {
		fake.processingEvents[id] = domain.RewardProcessingEvent{
			ID:            id,
			PartnerID:     "partner_1",
			MemberID:      "member_1",
			Type:          domain.EventPurchase,
			EligibleMinor: 10000,
			OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		}
	}
	fake.ruleGraphs = map[string]domain.RuleGraph{
		"version_1": {
			Groups: []domain.RuleGraphGroup{{ID: "base_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "base_rule", GroupID: "base_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 1}}},
		},
	}

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 2 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if fake.loadGraphsCalls != 1 {
		t.Fatalf("expected one batch graph load because second event uses cache, got %d", fake.loadGraphsCalls)
	}
}

func TestProcessFailureCreatesFailedCalculationAndMarksEventFailed(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.activeRuleSetErr = domain.InvariantError("member has no active enrollment")

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 0 || result.Failed != 1 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if fake.failedCalculations != 1 || fake.markedFailed != 1 {
		t.Fatalf("expected failed calculation and failed event status, got calculations=%d markedFailed=%d", fake.failedCalculations, fake.markedFailed)
	}
	if len(fake.createdCalculations) != 0 || len(fake.ledgerEntries) != 0 {
		t.Fatalf("expected no successful calculation or ledger entries, got calcs=%+v ledger=%+v", fake.createdCalculations, fake.ledgerEntries)
	}
}
