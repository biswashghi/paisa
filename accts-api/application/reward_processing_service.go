package application

import (
	"context"
	"fmt"

	"accts-api/domain"
	"accts-api/ports"
)

type RewardProcessingService struct {
	app app
}

func (s RewardProcessingService) ProcessTransactionEvents(ctx context.Context) (domain.ProcessTransactionEventsResult, error) {
	eventIDs, err := s.app.stores.Transactions.AcceptedIDs(ctx, 50)
	if err != nil {
		return domain.ProcessTransactionEventsResult{}, err
	}

	result := domain.ProcessTransactionEventsResult{}
	for _, eventID := range eventIDs {
		if err := s.processOneEvent(ctx, eventID); err != nil {
			result.Failed++
			_ = s.app.stores.Transactions.MarkFailed(ctx, eventID)
			continue
		}
		result.Processed++
	}
	return result, nil
}

func (s RewardProcessingService) processOneEvent(ctx context.Context, eventID string) error {
	return s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		if err := stores.Transactions.ClaimAccepted(ctx, eventID); err != nil {
			return domain.InvariantError(fmt.Sprintf("event not accepted: %s", eventID))
		}

		event, err := stores.Transactions.LoadForProcessing(ctx, eventID)
		if err != nil {
			return err
		}
		switch event.Type {
		case domain.EventRefund:
			_, err = s.processRefundEvent(ctx, stores, event)
		case domain.EventPurchase:
			_, err = s.processPurchaseEvent(ctx, stores, event)
		default:
			if err := stores.RewardCalculations.CreateFailed(ctx, event, fmt.Sprintf("unsupported transaction type %s", event.Type)); err != nil {
				return err
			}
			return stores.Transactions.MarkFailed(ctx, event.ID)
		}
		if err != nil {
			return err
		}
		return stores.Transactions.MarkProcessed(ctx, event.ID)
	})
}

func (s RewardProcessingService) processPurchaseEvent(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent) (string, error) {
	ruleSet, err := stores.Rules.ActiveProgramAndPublishedRuleSet(ctx, event.PartnerID, event.MemberID)
	if err != nil {
		return "", stores.RewardCalculations.CreateFailed(ctx, event, publicFailureReason(err))
	}
	calculation, err := s.calculateRuleSetPurchase(ctx, stores, event, ruleSet, true)
	if err != nil {
		return "", err
	}
	calculation["transactionType"] = event.Type
	calculation["ruleVersionId"] = ruleSet.BaseRuleVersionID
	calculation["ruleVersionIds"] = ruleSet.RuleVersionIDs
	calculation["baseRuleVersionId"] = ruleSet.BaseRuleVersionID
	calculation["addOnRuleVersionIds"] = ruleSet.AddOnRuleVersionIDs
	calculation["programId"] = ruleSet.ProgramID
	points := domain.IntFromAny(calculation["totalPoints"])

	calcID, err := stores.RewardCalculations.CreateSucceeded(ctx, ports.RewardCalculationCreateInput{
		PartnerID:          event.PartnerID,
		TransactionEventID: event.ID,
		ProgramID:          ruleSet.ProgramID,
		RuleVersionID:      ruleSet.BaseRuleVersionID,
		PointsDelta:        points,
		BasisAmountMinor:   event.EligibleMinor,
		CalculationData:    calculation,
	})
	if err != nil {
		return "", err
	}
	if points != 0 {
		accountID, err := stores.Members.AccountID(ctx, event.PartnerID, event.MemberID)
		if err != nil {
			return "", err
		}
		if _, err := postLedgerEntry(ctx, stores, ports.LedgerEntryInput{
			PartnerID:       event.PartnerID,
			MemberAccountID: accountID,
			ProgramID:       ruleSet.ProgramID,
			EntryType:       domain.EntryEarn,
			AvailableDelta:  points,
			SourceType:      "reward_calculation",
			SourceID:        calcID,
			Reason:          "purchase earn",
		}); err != nil {
			return "", err
		}
	}
	return calcID, nil
}

func (s RewardProcessingService) processRefundEvent(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent) (string, error) {
	programID, ruleVersionID, _ := stores.Rules.ActiveProgramAndPublishedRules(ctx, event.PartnerID, event.MemberID)
	refundCalc := domain.JSONMap{
		"transactionType":               event.Type,
		"originalExternalTransactionId": event.OriginalExternalTransactionID,
		"selectedAwards":                []domain.JSONMap{},
		"originalTransactionMissing":    false,
	}
	totalPoints := 0
	basis := event.EligibleMinor
	if basis <= 0 {
		basis = 1
	}
	calculationBasis := basis

	original, originalCalc, err := stores.RewardCalculations.OriginalForRefund(ctx, event)
	if err != nil {
		points, selectedAwards, fallbackErr := s.calculateMissingOriginalRefund(ctx, stores, event, programID, ruleVersionID, basis)
		if fallbackErr != nil {
			return "", stores.RewardCalculations.CreateFailed(ctx, event, publicFailureReason(fallbackErr))
		}
		totalPoints = points
		refundCalc["originalTransactionMissing"] = true
		refundCalc["selectedAwards"] = selectedAwards
	} else {
		priorRefundedBasis, err := stores.RewardCalculations.PriorRefundedBasisForOriginal(ctx, event)
		if err != nil {
			return "", err
		}
		remainingBasis := original.EligibleMinor - priorRefundedBasis
		if remainingBasis < 0 {
			remainingBasis = 0
		}
		effectiveBasis := domain.MinInt(basis, remainingBasis)
		calculationBasis = effectiveBasis
		ratioDenominator := original.EligibleMinor
		if ratioDenominator <= 0 {
			ratioDenominator = effectiveBasis
		}
		selectedAwards := domain.SelectedAwardsFromCalculation(originalCalc.CalculationData)
		reversed := []domain.JSONMap{}
		for _, award := range selectedAwards {
			reversedPoints := -domain.Prorate(award.Points, effectiveBasis, ratioDenominator)
			reversedBasis := domain.Prorate(award.BasisAmountMinor, effectiveBasis, ratioDenominator)
			totalPoints += reversedPoints
			reversed = append(reversed, domain.JSONMap{
				"ruleID":           award.RuleID,
				"ruleVersionID":    award.RuleVersionID,
				"basisAmountMinor": reversedBasis,
				"points":           reversedPoints,
			})
		}
		refundCalc["selectedAwards"] = reversed
		refundCalc["originalTransactionId"] = original.ID
		refundCalc["requestedBasisAmountMinor"] = basis
		refundCalc["effectiveBasisAmountMinor"] = effectiveBasis
		refundCalc["priorRefundedBasisAmountMinor"] = priorRefundedBasis
		refundCalc["remainingRefundableBasisAmountMinor"] = remainingBasis
		if programID == "" {
			programID = originalCalc.ProgramID
		}
		if ruleVersionID == "" {
			ruleVersionID = originalCalc.RuleVersionID
		}
	}
	refundCalc["totalPoints"] = totalPoints

	calcID, err := stores.RewardCalculations.CreateSucceeded(ctx, ports.RewardCalculationCreateInput{
		PartnerID:          event.PartnerID,
		TransactionEventID: event.ID,
		ProgramID:          programID,
		RuleVersionID:      ruleVersionID,
		PointsDelta:        totalPoints,
		BasisAmountMinor:   calculationBasis,
		CalculationData:    refundCalc,
	})
	if err != nil {
		return "", err
	}
	if totalPoints != 0 {
		accountID, err := stores.Members.AccountID(ctx, event.PartnerID, event.MemberID)
		if err != nil {
			return "", err
		}
		if _, err := postLedgerEntry(ctx, stores, ports.LedgerEntryInput{
			PartnerID:       event.PartnerID,
			MemberAccountID: accountID,
			ProgramID:       programID,
			EntryType:       domain.EntryRefund,
			AvailableDelta:  totalPoints,
			SourceType:      "reward_calculation",
			SourceID:        calcID,
			Reason:          "refund reversal",
		}); err != nil {
			return "", err
		}
	}
	return calcID, nil
}

func (s RewardProcessingService) calculateMissingOriginalRefund(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent, programID, ruleVersionID string, basis int) (int, []domain.JSONMap, error) {
	if ruleVersionID == "" {
		ruleSet, err := stores.Rules.ActiveProgramAndPublishedRuleSet(ctx, event.PartnerID, event.MemberID)
		if err != nil {
			return 0, nil, err
		}
		programID = ruleSet.ProgramID
		ruleVersionID = ruleSet.BaseRuleVersionID
		graphCalc, err := s.calculateRuleSetPurchase(ctx, stores, syntheticRefundPurchase(event, basis), ruleSet, false)
		if err != nil {
			return 0, nil, err
		}
		points := -domain.IntFromAny(graphCalc["totalPoints"])
		selected := []domain.JSONMap{}
		for _, award := range domain.SelectedAwardsFromCalculation(graphCalc) {
			selected = append(selected, domain.JSONMap{
				"ruleID":           award.RuleID,
				"ruleVersionID":    award.RuleVersionID,
				"basisAmountMinor": award.BasisAmountMinor,
				"points":           -award.Points,
				"fallback":         true,
			})
		}
		_ = programID
		_ = ruleVersionID
		return points, selected, nil
	}
	graph, err := stores.Rules.LoadGraph(ctx, ruleVersionID)
	if err != nil {
		return 0, nil, err
	}
	synthetic := event
	synthetic.Type = domain.EventPurchase
	synthetic.EligibleMinor = basis
	calculation, err := calculatePurchase(ctx, stores, synthetic, graph, ruleVersionID, false)
	if err != nil {
		return 0, nil, err
	}
	points := -domain.IntFromAny(calculation["totalPoints"])
	selected := []domain.JSONMap{}
	for _, award := range domain.SelectedAwardsFromCalculation(calculation) {
		selected = append(selected, domain.JSONMap{
			"ruleID":           award.RuleID,
			"ruleVersionID":    award.RuleVersionID,
			"basisAmountMinor": award.BasisAmountMinor,
			"points":           -award.Points,
			"fallback":         true,
		})
	}
	if len(selected) == 0 && points != 0 {
		selected = append(selected, domain.JSONMap{
			"ruleID":           "fallback_current_rules",
			"ruleVersionID":    ruleVersionID,
			"basisAmountMinor": basis,
			"points":           points,
			"fallback":         true,
		})
	}
	_ = programID
	return points, selected, nil
}

func (s RewardProcessingService) calculateRuleSetPurchase(ctx context.Context, stores ports.StoreSet, event domain.RewardProcessingEvent, ruleSet ports.RuleSetSelection, commitLimitUsage bool) (domain.JSONMap, error) {
	total := 0
	groups := []interface{}{}
	selectedAwards := []interface{}{}
	components := []domain.JSONMap{}
	for index, ruleVersionID := range ruleSet.RuleVersionIDs {
		graph, err := stores.Rules.LoadGraph(ctx, ruleVersionID)
		if err != nil {
			return nil, err
		}
		component, err := calculatePurchase(ctx, stores, event, graph, ruleVersionID, commitLimitUsage)
		if err != nil {
			return nil, err
		}
		total += domain.IntFromAny(component["totalPoints"])
		if rawGroups, ok := component["groups"].([]domain.JSONMap); ok {
			for _, group := range rawGroups {
				group["ruleVersionId"] = ruleVersionID
				groups = append(groups, group)
			}
		} else if rawGroups, ok := component["groups"].([]interface{}); ok {
			groups = append(groups, rawGroups...)
		}
		if rawAwards, ok := component["selectedAwards"].([]domain.JSONMap); ok {
			for _, award := range rawAwards {
				selectedAwards = append(selectedAwards, award)
			}
		} else if rawAwards, ok := component["selectedAwards"].([]interface{}); ok {
			selectedAwards = append(selectedAwards, rawAwards...)
		}
		scope := domain.RuleScopeMemberAddOn
		if index == 0 {
			scope = domain.RuleScopeProgramBase
		}
		components = append(components, domain.JSONMap{
			"ruleVersionId": ruleVersionID,
			"scope":         scope,
			"totalPoints":   domain.IntFromAny(component["totalPoints"]),
		})
	}
	return domain.JSONMap{
		"groups":         groups,
		"selectedAwards": selectedAwards,
		"components":     components,
		"totalPoints":    total,
	}, nil
}

func syntheticRefundPurchase(event domain.RewardProcessingEvent, basis int) domain.RewardProcessingEvent {
	synthetic := event
	synthetic.Type = domain.EventPurchase
	synthetic.EligibleMinor = basis
	return synthetic
}
