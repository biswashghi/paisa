package application

import (
	"context"
	"sort"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

type calcCandidate struct {
	RuleID           string `json:"ruleID"`
	RuleVersionID    string `json:"ruleVersionID"`
	BasisAmountMinor int    `json:"basisAmountMinor"`
	RawPoints        int    `json:"rawPoints"`
	Points           int    `json:"points"`
	Matched          bool   `json:"matched"`
	Selected         bool   `json:"selected"`
	Exhausted        bool   `json:"exhausted"`
	Reason           string `json:"reason"`
	LimitID          string `json:"limitID,omitempty"`
	UsagePoints      int    `json:"usagePoints"`
	UsageBasis       int    `json:"usageBasisAmountMinor"`
}

func calculatePurchase(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent, graph domain.RuleGraph, ruleVersionID string, commitLimitUsage bool) (domain.JSONMap, error) {
	total := 0
	traceGroups := []domain.JSONMap{}

	groups := append([]domain.RuleGraphGroup{}, graph.Groups...)
	sort.Slice(groups, func(i, j int) bool { return groups[i].Priority < groups[j].Priority })
	for _, group := range groups {
		groupRules := domain.RulesForGroup(graph.Rules, group.ID)
		var selected []calcCandidate
		var candidates []calcCandidate
		outcomes := map[string]calcCandidate{}
		switch group.Strategy {
		case "max_of":
			for _, rule := range groupRules {
				if ok, reason := dependenciesSatisfied(rule, graph.Dependencies, outcomes); !ok {
					candidate := dependencyBlockedCandidate(rule.ID, reason)
					candidates = append(candidates, candidate)
					outcomes[rule.ID] = candidate
					continue
				}
				candidate, err := evaluateRuleCandidate(ctx, stores, event, rule, 0, false)
				if err != nil {
					return nil, err
				}
				candidates = append(candidates, candidate)
				outcomes[rule.ID] = candidate
			}
			best := -1
			for i, candidate := range candidates {
				if candidate.Points == 0 {
					continue
				}
				if best == -1 || candidate.Points > candidates[best].Points {
					best = i
				}
			}
			if best >= 0 {
				candidates[best].Selected = true
				candidates[best].Reason = "highest_candidate_after_limits"
				selected = append(selected, candidates[best])
				outcomes[candidates[best].RuleID] = candidates[best]
			}
		case "waterfall":
			remaining := basisByCategory(event)
			for _, rule := range groupRules {
				if ok, reason := dependenciesSatisfied(rule, graph.Dependencies, outcomes); !ok {
					candidate := dependencyBlockedCandidate(rule.ID, reason)
					candidates = append(candidates, candidate)
					outcomes[rule.ID] = candidate
					continue
				}
				basis := waterfallBasis(event, rule, remaining)
				candidate, err := evaluateRuleCandidate(ctx, stores, event, rule, basis, true)
				if err != nil {
					return nil, err
				}
				if candidate.Points != 0 {
					candidate.Selected = true
					candidate.Reason = "waterfall_selected"
					consumeWaterfallBasis(rule, remaining, candidate.BasisAmountMinor)
					selected = append(selected, candidate)
				}
				candidates = append(candidates, candidate)
				outcomes[rule.ID] = candidate
			}
		default:
			for _, rule := range groupRules {
				if ok, reason := dependenciesSatisfied(rule, graph.Dependencies, outcomes); !ok {
					candidate := dependencyBlockedCandidate(rule.ID, reason)
					candidates = append(candidates, candidate)
					outcomes[rule.ID] = candidate
					continue
				}
				candidate, err := evaluateRuleCandidate(ctx, stores, event, rule, 0, false)
				if err != nil {
					return nil, err
				}
				if candidate.Points != 0 {
					candidate.Selected = true
					candidate.Reason = "stack_selected"
					selected = append(selected, candidate)
				}
				candidates = append(candidates, candidate)
				outcomes[rule.ID] = candidate
			}
		}

		for _, candidate := range selected {
			delta := ports.RuleLimitUsageDelta{
				LimitID:     candidate.LimitID,
				UsagePoints: candidate.UsagePoints,
				UsageBasis:  candidate.UsageBasis,
			}
			if commitLimitUsage {
				if err := stores.Rules.CommitLimitUsage(ctx, event.PartnerID, event.MemberID, event.OccurredAt, delta); err != nil {
					return nil, err
				}
			}
			total += candidate.Points
		}
		traceGroups = append(traceGroups, domain.JSONMap{"ruleGroupID": group.ID, "strategy": group.Strategy, "candidates": candidates})
	}

	selectedAwards := []domain.JSONMap{}
	for _, group := range traceGroups {
		for _, value := range group["candidates"].([]calcCandidate) {
			if value.Selected {
				selectedAwards = append(selectedAwards, domain.JSONMap{
					"ruleID":           value.RuleID,
					"ruleVersionID":    ruleVersionID,
					"basisAmountMinor": value.BasisAmountMinor,
					"points":           value.Points,
				})
			}
		}
	}
	return domain.JSONMap{"groups": traceGroups, "selectedAwards": selectedAwards, "totalPoints": total}, nil
}

func evaluateRuleCandidate(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent, rule domain.RuleGraphRule, forcedBasis int, useForcedBasis bool) (calcCandidate, error) {
	candidate := calcCandidate{RuleID: rule.ID}
	eligible, err := eligibleForRule(ctx, stores, event, rule)
	if err != nil {
		return candidate, err
	}
	if !eligible {
		candidate.Reason = "not_eligible"
		return candidate, nil
	}
	candidate.Matched = true
	basis := forcedBasis
	if !useForcedBasis {
		basis = basisForRule(event, rule)
	}
	candidate.BasisAmountMinor = basis
	points := 0
	switch rule.RuleType {
	case domain.RuleTypeFixedPerTransaction, domain.RuleTypeFirstPurchaseBonus, domain.RuleTypeSpendWindowBonus:
		points = domain.ConfigInt(rule.Formula, "points", "fixedPoints")
	default:
		pointsPerDollar := domain.ConfigInt(rule.Formula, "points_per_dollar", "pointsPerDollar")
		points = (basis / 100) * pointsPerDollar
	}
	candidate.RawPoints = points
	candidate.Points = points

	limits, err := stores.Rules.LimitsForRule(ctx, rule.ID)
	if err != nil {
		return candidate, err
	}
	for _, limit := range limits {
		candidate.LimitID = limit.ID
		usedPoints, usedBasis, err := stores.Rules.CurrentLimitUsage(ctx, event.MemberID, limit, event.OccurredAt)
		if err != nil {
			return candidate, err
		}
		if limit.MaxBasisAmountMinor > 0 {
			remainingBasis := limit.MaxBasisAmountMinor - usedBasis
			if remainingBasis < 0 {
				remainingBasis = 0
			}
			if candidate.BasisAmountMinor > remainingBasis {
				candidate.Exhausted = true
				candidate.BasisAmountMinor = remainingBasis
				pointsPerDollar := domain.ConfigInt(rule.Formula, "points_per_dollar", "pointsPerDollar")
				if pointsPerDollar == 0 {
					pointsPerDollar = 1
				}
				candidate.Points = (remainingBasis / 100) * pointsPerDollar
			}
			candidate.UsageBasis = candidate.BasisAmountMinor
			candidate.UsagePoints = candidate.Points
		}
		if limit.MaxPoints > 0 {
			remainingPoints := limit.MaxPoints - usedPoints
			if remainingPoints < 0 {
				remainingPoints = 0
			}
			if candidate.Points > remainingPoints {
				candidate.Exhausted = true
				candidate.Points = remainingPoints
			}
			candidate.UsagePoints = candidate.Points
		}
	}
	return candidate, nil
}

func eligibleForRule(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent, rule domain.RuleGraphRule) (bool, error) {
	transactionTypes := domain.StringSliceConfig(rule.Eligibility, "transaction_types")
	if len(transactionTypes) == 0 {
		transactionTypes = domain.StringSliceConfig(rule.Eligibility, "transactionTypes")
	}
	if len(transactionTypes) > 0 && !domain.ContainsString(transactionTypes, event.Type) {
		return false, nil
	}
	if rule.RuleType == domain.RuleTypeFirstPurchaseBonus || domain.BoolConfig(rule.Eligibility, "firstPurchase") {
		count, err := stores.Transactions.PriorProcessedPurchaseCount(ctx, event.PartnerID, event.MemberID, event.ID)
		if err != nil {
			return false, err
		}
		if count > 0 {
			return false, nil
		}
	}
	if channels := domain.StringSliceConfig(rule.Eligibility, "channels"); len(channels) > 0 {
		channel := domain.StringFromAny(event.RawPayload["channel"])
		if channel == "" || !domain.ContainsString(channels, channel) {
			return false, nil
		}
	}
	if rule.RuleType == domain.RuleTypeSpendWindowBonus {
		windowDays := domain.ConfigInt(rule.Eligibility, "window_days", "windowDays")
		threshold := domain.ConfigInt(rule.Eligibility, "spend_threshold_minor", "spendThresholdMinor", "thresholdMinor")
		since := event.OccurredAt.Add(-time.Duration(windowDays) * 24 * time.Hour)
		prior, err := stores.Transactions.PriorProcessedPurchaseEligibleMinorSum(ctx, event.PartnerID, event.MemberID, event.ID, since)
		if err != nil {
			return false, err
		}
		if prior+event.EligibleMinor < threshold {
			return false, nil
		}
		return true, nil
	}
	return basisForRule(event, rule) > 0 || rule.RuleType == domain.RuleTypeFixedPerTransaction || rule.RuleType == domain.RuleTypeFirstPurchaseBonus, nil
}

func basisForRule(event domain.RewardProcessingEvent, rule domain.RuleGraphRule) int {
	basisName := domain.StringFromAny(rule.Eligibility["basis"])
	if basisName == "" {
		basisName = "eligible"
	}
	categories := domain.StringSliceConfig(rule.Eligibility, "categories")
	if len(categories) > 0 {
		total := 0
		for _, line := range event.Lines {
			if domain.ContainsString(categories, line.Category) {
				total += lineBasis(line, basisName)
			}
		}
		return total
	}
	switch basisName {
	case "subtotal":
		return event.SubtotalMinor
	case "tax":
		return event.TaxMinor
	case "total":
		return event.TotalMinor
	case "line_item_subtotal", "lineItemSubtotal", "line_item_total", "lineItemTotal", "line_item_eligible", "lineItemEligible":
		total := 0
		for _, line := range event.Lines {
			total += lineBasis(line, basisName)
		}
		return total
	default:
		return event.EligibleMinor
	}
}

func basisByCategory(event domain.RewardProcessingEvent) map[string]int {
	out := map[string]int{}
	for _, line := range event.Lines {
		out[line.Category] += line.EligibleMinor
	}
	if len(out) == 0 {
		out[""] = event.EligibleMinor
	}
	return out
}

func waterfallBasis(event domain.RewardProcessingEvent, rule domain.RuleGraphRule, remaining map[string]int) int {
	categories := domain.StringSliceConfig(rule.Eligibility, "categories")
	if len(categories) == 0 {
		total := 0
		for _, value := range remaining {
			total += value
		}
		return total
	}
	total := 0
	for _, category := range categories {
		total += remaining[category]
	}
	return total
}

func consumeWaterfallBasis(rule domain.RuleGraphRule, remaining map[string]int, amount int) {
	categories := domain.StringSliceConfig(rule.Eligibility, "categories")
	if len(categories) == 0 {
		categories = []string{""}
	}
	for _, category := range categories {
		if amount <= 0 {
			return
		}
		consume := domain.MinInt(remaining[category], amount)
		remaining[category] -= consume
		amount -= consume
	}
}

func dependencyBlockedCandidate(ruleID, reason string) calcCandidate {
	return calcCandidate{RuleID: ruleID, Reason: reason}
}

func dependenciesSatisfied(rule domain.RuleGraphRule, deps []domain.RuleGraphDependency, outcomes map[string]calcCandidate) (bool, string) {
	for _, dep := range deps {
		if dep.RuleID != rule.ID {
			continue
		}
		prior, ok := outcomes[dep.DependsOnRuleID]
		if !ok {
			return false, "dependency_not_evaluated"
		}
		switch dep.DependencyType {
		case "requires_match":
			if !prior.Matched {
				return false, "dependency_not_matched"
			}
		case "requires_award":
			if !prior.Selected || prior.Points == 0 {
				return false, "dependency_not_awarded"
			}
		case "requires_exhausted":
			if !prior.Exhausted {
				return false, "dependency_not_exhausted"
			}
		case "blocked_if_awarded":
			if prior.Selected && prior.Points != 0 {
				return false, "dependency_blocked_by_award"
			}
		case "applies_to_remainder":
			continue
		}
	}
	return true, ""
}

func lineBasis(line domain.LineItemInput, basisName string) int {
	switch basisName {
	case "subtotal", "line_item_subtotal", "lineItemSubtotal":
		return line.SubtotalMinor
	case "tax":
		return line.TaxMinor
	case "total", "line_item_total", "lineItemTotal":
		return line.TotalMinor
	default:
		return line.EligibleMinor
	}
}
