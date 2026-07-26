package application

import (
	"context"
	"testing"

	"accts-api/domain"
)

func TestPublishRuleVersionRejectsInvalidRuleGraph(t *testing.T) {
	fake := newFakeAppStore()
	fake.ruleVersionStatus = domain.RuleDraft
	fake.ruleGraph = domain.RuleGraph{
		Groups: []domain.RuleGraphGroup{{ID: "group_1", Strategy: "max_of"}},
		Rules:  []domain.RuleGraphRule{{ID: "rule_1", GroupID: "group_1"}},
	}

	_, err := ProgramService{app: testApp(fake)}.PublishRuleVersion(context.Background(), "acme-demo", "program_1", "version_1")
	if !domain.IsErrorKind(err, domain.ErrorKindInvalid) {
		t.Fatalf("expected invalid rule graph error, got %v", err)
	}
	if fake.publishedRuleVersion {
		t.Fatal("expected invalid rule graph to prevent publishing")
	}
}

func TestPublishRuleVersionRejectsUnsupportedDependency(t *testing.T) {
	fake := newFakeAppStore()
	fake.ruleVersionStatus = domain.RuleDraft
	fake.ruleGraph = domain.RuleGraph{
		Groups: []domain.RuleGraphGroup{{ID: "group_1", Strategy: domain.RuleStrategyStack}},
		Rules: []domain.RuleGraphRule{
			{ID: "rule_1", GroupID: "group_1", RuleType: domain.RuleTypePointsPerDollar, Priority: 1, Formula: domain.JSONMap{"pointsPerDollar": 1}},
			{ID: "rule_2", GroupID: "group_1", RuleType: domain.RuleTypePointsPerDollar, Priority: 2, Formula: domain.JSONMap{"pointsPerDollar": 1}},
		},
		Dependencies: []domain.RuleGraphDependency{{RuleID: "rule_2", DependsOnRuleID: "rule_1", DependencyType: "mystery"}},
	}

	_, err := ProgramService{app: testApp(fake)}.PublishRuleVersion(context.Background(), "acme-demo", "program_1", "version_1")
	if !domain.IsErrorKind(err, domain.ErrorKindInvalid) {
		t.Fatalf("expected invalid dependency error, got %v", err)
	}
}
