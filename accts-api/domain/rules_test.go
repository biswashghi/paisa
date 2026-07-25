package domain

import "testing"

func TestValidateNoDependencyCyclesRejectsCycle(t *testing.T) {
	rules := map[string]RuleGraphRule{
		"rule_a": {ID: "rule_a"},
		"rule_b": {ID: "rule_b"},
	}
	deps := []RuleGraphDependency{
		{RuleID: "rule_a", DependsOnRuleID: "rule_b"},
		{RuleID: "rule_b", DependsOnRuleID: "rule_a"},
	}

	if err := ValidateNoDependencyCycles(rules, deps); err == nil {
		t.Fatal("expected dependency cycle to be rejected")
	}
}

func TestHashTransactionPayloadIsStableAndDetectsChangedPayload(t *testing.T) {
	original := TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  EventPurchase,
		Currency:              "USD",
		EligibleMinor:         10000,
	}
	retry := original
	changed := original
	changed.EligibleMinor = 20000

	if HashTransactionPayload(original) != HashTransactionPayload(retry) {
		t.Fatal("expected identical payloads to hash the same")
	}
	if HashTransactionPayload(original) == HashTransactionPayload(changed) {
		t.Fatal("expected changed payload to produce a different hash")
	}
}

func TestSelectedAwardsFromCalculation(t *testing.T) {
	awards := SelectedAwardsFromCalculation(JSONMap{
		"selectedAwards": []interface{}{
			map[string]interface{}{
				"ruleID":           "base",
				"ruleVersionID":    "rules_v1",
				"basisAmountMinor": float64(10000),
				"points":           float64(100),
			},
		},
	})

	if len(awards) != 1 {
		t.Fatalf("expected one award, got %d", len(awards))
	}
	if awards[0].RuleID != "base" || awards[0].RuleVersionID != "rules_v1" || awards[0].BasisAmountMinor != 10000 || awards[0].Points != 100 {
		t.Fatalf("unexpected award parse: %+v", awards[0])
	}
}

func TestProrate(t *testing.T) {
	if got := Prorate(100, 5000, 10000); got != 50 {
		t.Fatalf("expected half proration, got %d", got)
	}
	if got := Prorate(100, 1, 0); got != 0 {
		t.Fatalf("expected zero denominator to return 0, got %d", got)
	}
}
