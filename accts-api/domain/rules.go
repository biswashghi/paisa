package domain

import (
	"fmt"
	"sort"
)

type RuleGraphGroup struct {
	ID       string
	Strategy string
	Priority int
}

type RuleGraphRule struct {
	ID          string
	GroupID     string
	RuleKey     string
	RuleType    string
	Priority    int
	Eligibility JSONMap
	Formula     JSONMap
}

type RuleGraphLimit struct {
	ID                  string
	RuleID              string
	Scope               string
	Period              string
	MaxPoints           int
	MaxBasisAmountMinor int
}

type RuleGraphDependency struct {
	RuleID          string
	DependsOnRuleID string
	DependencyType  string
}

type RuleGraph struct {
	Groups       []RuleGraphGroup
	Rules        []RuleGraphRule
	Limits       []RuleGraphLimit
	Dependencies []RuleGraphDependency
}

type SelectedAward struct {
	RuleID           string
	RuleVersionID    string
	BasisAmountMinor int
	Points           int
}

func RulesForGroup(rules []RuleGraphRule, groupID string) []RuleGraphRule {
	out := []RuleGraphRule{}
	for _, rule := range rules {
		if rule.GroupID == groupID {
			out = append(out, rule)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	return out
}

func DependencyForRule(deps []RuleGraphDependency, ruleID, dependencyType string) *RuleGraphDependency {
	for _, dep := range deps {
		if dep.RuleID == ruleID && dep.DependencyType == dependencyType {
			return &dep
		}
	}
	return nil
}

func ValidateRuleGraphForPublish(groups []RuleGraphGroup, rules []RuleGraphRule, deps []RuleGraphDependency, limits ...[]RuleGraphLimit) error {
	if len(rules) == 0 {
		return InvalidError("rule version has no active rules")
	}
	if len(groups) == 0 {
		return InvalidError("rule version has no active groups")
	}

	rulesByID := map[string]RuleGraphRule{}
	groupsByID := map[string]RuleGraphGroup{}
	for _, rule := range rules {
		if !SupportedRuleType(rule.RuleType) {
			return InvalidError(fmt.Sprintf("unsupported rule type %s", rule.RuleType))
		}
		if err := validateRuleConfig(rule); err != nil {
			return err
		}
		rulesByID[rule.ID] = rule
	}
	for _, group := range groups {
		if !SupportedRuleStrategy(group.Strategy) {
			return InvalidError(fmt.Sprintf("unsupported resolution strategy %s", group.Strategy))
		}
		groupsByID[group.ID] = group
		activeCount := 0
		for _, rule := range rules {
			if rule.GroupID == group.ID {
				activeCount++
			}
		}
		if group.Strategy == "max_of" && activeCount < 2 {
			return InvalidError(fmt.Sprintf("max_of group %s must have at least two active rules", group.ID))
		}
	}
	if err := validateDependencies(rulesByID, groupsByID, deps); err != nil {
		return err
	}
	if len(limits) > 0 {
		if err := ValidateRuleLimitsForPublish(limits[0]); err != nil {
			return err
		}
	}
	return ValidateNoDependencyCycles(rulesByID, deps)
}

func ValidateRuleLimitsForPublish(limits []RuleGraphLimit) error {
	for _, limit := range limits {
		if !SupportedLimitScope(limit.Scope) {
			return InvalidError(fmt.Sprintf("unsupported limit scope %s", limit.Scope))
		}
		if !SupportedLimitPeriod(limit.Period) {
			return InvalidError(fmt.Sprintf("unsupported limit period %s", limit.Period))
		}
		if limit.MaxPoints <= 0 && limit.MaxBasisAmountMinor <= 0 {
			return InvalidError("rule limit must set maxPoints or maxBasisAmountMinor")
		}
	}
	return nil
}

func validateRuleConfig(rule RuleGraphRule) error {
	transactionTypes := StringSliceConfig(rule.Eligibility, "transaction_types")
	if len(transactionTypes) == 0 {
		transactionTypes = StringSliceConfig(rule.Eligibility, "transactionTypes")
	}
	for _, eventType := range transactionTypes {
		if eventType != EventPurchase {
			return InvalidError("earning rules can only target purchase transaction types")
		}
	}
	switch rule.RuleType {
	case RuleTypePointsPerDollar:
		if ConfigInt(rule.Formula, "points_per_dollar", "pointsPerDollar") <= 0 {
			return InvalidError(fmt.Sprintf("rule %s must set pointsPerDollar", rule.ID))
		}
	case RuleTypeFixedPerTransaction, RuleTypeFirstPurchaseBonus:
		if ConfigInt(rule.Formula, "points", "fixedPoints") <= 0 {
			return InvalidError(fmt.Sprintf("rule %s must set points", rule.ID))
		}
	case RuleTypeSpendWindowBonus:
		if ConfigInt(rule.Formula, "points", "fixedPoints") <= 0 {
			return InvalidError(fmt.Sprintf("rule %s must set points", rule.ID))
		}
		if ConfigInt(rule.Eligibility, "window_days", "windowDays") <= 0 {
			return InvalidError(fmt.Sprintf("rule %s must set windowDays", rule.ID))
		}
		if ConfigInt(rule.Eligibility, "spend_threshold_minor", "spendThresholdMinor", "thresholdMinor") <= 0 {
			return InvalidError(fmt.Sprintf("rule %s must set spendThresholdMinor", rule.ID))
		}
	}
	return nil
}

func validateDependencies(rules map[string]RuleGraphRule, groups map[string]RuleGraphGroup, deps []RuleGraphDependency) error {
	for _, dep := range deps {
		rule, ok := rules[dep.RuleID]
		if !ok {
			continue
		}
		dependsOn, ok := rules[dep.DependsOnRuleID]
		if !ok {
			return InvalidError("dependency references unknown or disabled rule")
		}
		if rule.GroupID != dependsOn.GroupID {
			return InvalidError("dependencies must stay within the same rule group")
		}
		if dependsOn.Priority >= rule.Priority {
			return InvalidError("dependency must point backward")
		}
		group := groups[rule.GroupID]
		switch dep.DependencyType {
		case "requires_match", "requires_award", "requires_exhausted", "blocked_if_awarded":
		case "applies_to_remainder":
			if group.Strategy != RuleStrategyWaterfall {
				return InvalidError("applies_to_remainder is only supported in waterfall groups")
			}
		default:
			return InvalidError(fmt.Sprintf("unsupported dependency type %s", dep.DependencyType))
		}
		if group.Strategy == RuleStrategyMaxOf && dep.DependencyType != "requires_match" {
			return InvalidError("max_of groups only support requires_match dependencies")
		}
	}
	return nil
}

func ValidateNoDependencyCycles(rules map[string]RuleGraphRule, deps []RuleGraphDependency) error {
	graph := map[string][]string{}
	for _, dep := range deps {
		if _, ok := rules[dep.RuleID]; !ok {
			continue
		}
		if _, ok := rules[dep.DependsOnRuleID]; !ok {
			return InvalidError("dependency references unknown or disabled rule")
		}
		graph[dep.RuleID] = append(graph[dep.RuleID], dep.DependsOnRuleID)
	}

	state := map[string]int{}
	var visit func(string) error
	visit = func(ruleID string) error {
		switch state[ruleID] {
		case 1:
			return InvalidError("dependency cycle detected")
		case 2:
			return nil
		}
		state[ruleID] = 1
		for _, next := range graph[ruleID] {
			if err := visit(next); err != nil {
				return err
			}
		}
		state[ruleID] = 2
		return nil
	}
	for ruleID := range rules {
		if err := visit(ruleID); err != nil {
			return err
		}
	}
	return nil
}

func SelectedAwardsFromCalculation(data JSONMap) []SelectedAward {
	awards := []SelectedAward{}
	switch raw := data["selectedAwards"].(type) {
	case []interface{}:
		for _, item := range raw {
			if m, ok := item.(map[string]interface{}); ok {
				awards = append(awards, awardFromMap(m))
			}
		}
	case []JSONMap:
		for _, item := range raw {
			awards = append(awards, awardFromMap(item))
		}
	case []map[string]interface{}:
		for _, item := range raw {
			awards = append(awards, awardFromMap(item))
		}
	}
	return awards
}

func awardFromMap(m map[string]interface{}) SelectedAward {
	return SelectedAward{
		RuleID:           StringFromAny(m["ruleID"]),
		RuleVersionID:    StringFromAny(m["ruleVersionID"]),
		BasisAmountMinor: IntFromAny(m["basisAmountMinor"]),
		Points:           IntFromAny(m["points"]),
	}
}
