package repository

import (
	"context"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

func (s RuleStore) LoadGraph(ctx context.Context, ruleVersionID string) (domain.RuleGraph, error) {
	groupRows, err := s.q.QueryContext(ctx, `
		SELECT id, resolution_strategy, priority
		FROM paisa.rule_groups
		WHERE rule_version_id = $1 AND status = 'active'
		ORDER BY priority`, ruleVersionID)
	if err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}
	defer groupRows.Close()
	groups := []domain.RuleGraphGroup{}
	for groupRows.Next() {
		var group domain.RuleGraphGroup
		if err := groupRows.Scan(&group.ID, &group.Strategy, &group.Priority); err != nil {
			return domain.RuleGraph{}, AppErrorFromDB(err)
		}
		groups = append(groups, group)
	}
	if err := groupRows.Err(); err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}

	ruleRows, err := s.q.QueryContext(ctx, `
		SELECT er.id, er.rule_group_id, er.rule_key, er.rule_type, er.priority, er.eligibility_config, er.formula_config
		FROM paisa.earning_rules er
		JOIN paisa.rule_groups rg ON rg.id = er.rule_group_id
		WHERE rg.rule_version_id = $1 AND er.status = 'active'
		ORDER BY er.priority`, ruleVersionID)
	if err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}
	defer ruleRows.Close()
	rules := []domain.RuleGraphRule{}
	for ruleRows.Next() {
		var rule domain.RuleGraphRule
		var eligibility, formula []byte
		if err := ruleRows.Scan(&rule.ID, &rule.GroupID, &rule.RuleKey, &rule.RuleType, &rule.Priority, &eligibility, &formula); err != nil {
			return domain.RuleGraph{}, AppErrorFromDB(err)
		}
		rule.Eligibility = scanJSON(eligibility)
		rule.Formula = scanJSON(formula)
		rules = append(rules, rule)
	}
	if err := ruleRows.Err(); err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}

	limitRows, err := s.q.QueryContext(ctx, `
		SELECT rl.id, rl.rule_id, rl.scope, rl.period, COALESCE(rl.max_points, 0), COALESCE(rl.max_basis_amount_minor, 0)
		FROM paisa.rule_limits rl
		JOIN paisa.earning_rules er ON er.id = rl.rule_id
		JOIN paisa.rule_groups rg ON rg.id = er.rule_group_id
		WHERE rg.rule_version_id = $1 AND rl.status = 'active'`, ruleVersionID)
	if err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}
	defer limitRows.Close()
	limits := []domain.RuleGraphLimit{}
	for limitRows.Next() {
		var limit domain.RuleGraphLimit
		if err := limitRows.Scan(&limit.ID, &limit.RuleID, &limit.Scope, &limit.Period, &limit.MaxPoints, &limit.MaxBasisAmountMinor); err != nil {
			return domain.RuleGraph{}, AppErrorFromDB(err)
		}
		limits = append(limits, limit)
	}
	if err := limitRows.Err(); err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}

	depRows, err := s.q.QueryContext(ctx, `
		SELECT rd.rule_id, rd.depends_on_rule_id, rd.dependency_type
		FROM paisa.rule_dependencies rd
		JOIN paisa.earning_rules er ON er.id = rd.rule_id
		JOIN paisa.rule_groups rg ON rg.id = er.rule_group_id
		WHERE rg.rule_version_id = $1`, ruleVersionID)
	if err != nil {
		return domain.RuleGraph{}, AppErrorFromDB(err)
	}
	defer depRows.Close()
	deps := []domain.RuleGraphDependency{}
	for depRows.Next() {
		var dep domain.RuleGraphDependency
		if err := depRows.Scan(&dep.RuleID, &dep.DependsOnRuleID, &dep.DependencyType); err != nil {
			return domain.RuleGraph{}, AppErrorFromDB(err)
		}
		deps = append(deps, dep)
	}
	return domain.RuleGraph{Groups: groups, Rules: rules, Limits: limits, Dependencies: deps}, AppErrorFromDB(depRows.Err())
}

func (s RuleStore) LimitsForRule(ctx context.Context, ruleID string) ([]domain.RuleGraphLimit, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, rule_id, scope, period, COALESCE(max_points, 0), COALESCE(max_basis_amount_minor, 0)
		FROM paisa.rule_limits
		WHERE rule_id = $1 AND status = 'active'`, ruleID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	limits := []domain.RuleGraphLimit{}
	for rows.Next() {
		var limit domain.RuleGraphLimit
		if err := rows.Scan(&limit.ID, &limit.RuleID, &limit.Scope, &limit.Period, &limit.MaxPoints, &limit.MaxBasisAmountMinor); err != nil {
			return nil, AppErrorFromDB(err)
		}
		limits = append(limits, limit)
	}
	return limits, AppErrorFromDB(rows.Err())
}

func (s RuleStore) CurrentLimitUsage(ctx context.Context, memberID string, limit domain.RuleGraphLimit, occurredAt time.Time) (int, int, error) {
	periodKey := domain.PeriodKey(limit.Period, occurredAt)
	var usedPoints, usedBasis int
	err := s.q.QueryRowContext(ctx, `
		SELECT used_points, used_basis_amount_minor
		FROM paisa.rule_limit_usage
		WHERE rule_limit_id = $1 AND member_id = $2 AND period_key = $3
		FOR UPDATE`, limit.ID, memberID, periodKey,
	).Scan(&usedPoints, &usedBasis)
	if isNotFound(err) {
		return 0, 0, nil
	}
	return usedPoints, usedBasis, AppErrorFromDB(err)
}

func (s RuleStore) CommitLimitUsage(ctx context.Context, partnerID, memberID string, occurredAt time.Time, delta ports.RuleLimitUsageDelta) error {
	if delta.LimitID == "" {
		return nil
	}
	period := "lifetime"
	var ruleID string
	if err := s.q.QueryRowContext(ctx, `SELECT rule_id, period FROM paisa.rule_limits WHERE id = $1`, delta.LimitID).Scan(&ruleID, &period); err != nil {
		return AppErrorFromDB(err)
	}
	_ = ruleID
	key := domain.PeriodKey(period, occurredAt)
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO paisa.rule_limit_usage (partner_id, rule_limit_id, member_id, period_key, used_points, used_basis_amount_minor)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (rule_limit_id, member_id, period_key)
		DO UPDATE SET used_points = paisa.rule_limit_usage.used_points + EXCLUDED.used_points,
			used_basis_amount_minor = paisa.rule_limit_usage.used_basis_amount_minor + EXCLUDED.used_basis_amount_minor,
			updated_at = now()`,
		partnerID, delta.LimitID, memberID, key, delta.UsagePoints, delta.UsageBasis)
	return AppErrorFromDB(err)
}

func (s RuleStore) ActiveProgramAndPublishedRules(ctx context.Context, partnerID, memberID string) (string, string, error) {
	var programID string
	err := s.q.QueryRowContext(ctx, `
		SELECT pe.program_id
		FROM paisa.program_enrollments pe
		JOIN paisa.partners p ON p.id = pe.partner_id
		JOIN paisa.members m ON m.id = pe.member_id
		JOIN paisa.member_accounts ma ON ma.member_id = m.id
		JOIN paisa.programs pr ON pr.id = pe.program_id
		WHERE pe.partner_id = $1
			AND pe.member_id = $2
			AND pe.status = 'active'
			AND p.status = 'active'
			AND m.status = 'active'
			AND ma.status = 'active'
			AND pr.status = 'active'`, partnerID, memberID,
	).Scan(&programID)
	if err != nil {
		return "", "", AppErrorFromDB(err)
	}
	var ruleVersionID string
	err = s.q.QueryRowContext(ctx, `
		SELECT id
		FROM paisa.program_rule_versions
		WHERE partner_id = $1 AND program_id = $2 AND status = $3 AND scope = $4
		ORDER BY published_at DESC
		LIMIT 1`, partnerID, programID, domain.RulePublished, domain.RuleScopeProgramBase,
	).Scan(&ruleVersionID)
	return programID, ruleVersionID, AppErrorFromDB(err)
}

func (s RuleStore) ActiveProgramAndPublishedRuleSet(ctx context.Context, partnerID, memberID string) (ports.RuleSetSelection, error) {
	programID, baseRuleVersionID, err := s.ActiveProgramAndPublishedRules(ctx, partnerID, memberID)
	if err != nil {
		return ports.RuleSetSelection{}, err
	}
	rows, err := s.q.QueryContext(ctx, `
		SELECT mra.rule_version_id
		FROM paisa.member_rule_assignments mra
		JOIN paisa.program_enrollments pe ON pe.id = mra.program_enrollment_id
		JOIN paisa.program_rule_versions prv ON prv.id = mra.rule_version_id
		WHERE mra.partner_id = $1
			AND mra.member_id = $2
			AND mra.status = 'active'
			AND pe.status = 'active'
			AND pe.program_id = $3
			AND prv.status = $4
			AND prv.scope = $5
			AND mra.starts_at <= now()
			AND (mra.ends_at IS NULL OR mra.ends_at > now())
		ORDER BY mra.starts_at, mra.created_at`, partnerID, memberID, programID, domain.RulePublished, domain.RuleScopeMemberAddOn)
	if err != nil {
		return ports.RuleSetSelection{}, AppErrorFromDB(err)
	}
	defer rows.Close()
	addOns := []string{}
	for rows.Next() {
		var ruleVersionID string
		if err := rows.Scan(&ruleVersionID); err != nil {
			return ports.RuleSetSelection{}, AppErrorFromDB(err)
		}
		addOns = append(addOns, ruleVersionID)
	}
	if err := rows.Err(); err != nil {
		return ports.RuleSetSelection{}, AppErrorFromDB(err)
	}
	all := append([]string{baseRuleVersionID}, addOns...)
	return ports.RuleSetSelection{
		ProgramID:           programID,
		BaseRuleVersionID:   baseRuleVersionID,
		AddOnRuleVersionIDs: addOns,
		RuleVersionIDs:      all,
	}, nil
}

func (s RuleStore) ReviewGroups(ctx context.Context, ruleVersionID string) ([]domain.RuleGroupReview, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, name, resolution_strategy, priority, status
		FROM paisa.rule_groups
		WHERE rule_version_id = $1
		ORDER BY priority, created_at`, ruleVersionID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	groups := []domain.RuleGroupReview{}
	for rows.Next() {
		var group domain.RuleGroupReview
		if err := rows.Scan(&group.ID, &group.Name, &group.ResolutionStrategy, &group.Priority, &group.Status); err != nil {
			return nil, AppErrorFromDB(err)
		}
		rules, err := s.earningRules(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		group.Rules = rules
		groups = append(groups, group)
	}
	return groups, AppErrorFromDB(rows.Err())
}

func (s RuleStore) earningRules(ctx context.Context, groupID string) ([]domain.EarningRuleReview, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, rule_key, name, rule_type, priority, status, eligibility_config, formula_config
		FROM paisa.earning_rules
		WHERE rule_group_id = $1
		ORDER BY priority, created_at`, groupID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	rules := []domain.EarningRuleReview{}
	for rows.Next() {
		var rule domain.EarningRuleReview
		var eligibility, formula []byte
		if err := rows.Scan(&rule.ID, &rule.RuleKey, &rule.Name, &rule.RuleType, &rule.Priority, &rule.Status, &eligibility, &formula); err != nil {
			return nil, AppErrorFromDB(err)
		}
		rule.EligibilityConfig = scanJSON(eligibility)
		rule.FormulaConfig = scanJSON(formula)
		limits, err := s.ruleLimits(ctx, rule.ID)
		if err != nil {
			return nil, err
		}
		deps, err := s.ruleDependencies(ctx, rule.ID)
		if err != nil {
			return nil, err
		}
		rule.Limits = limits
		rule.Dependencies = deps
		rules = append(rules, rule)
	}
	return rules, AppErrorFromDB(rows.Err())
}

func (s RuleStore) ruleLimits(ctx context.Context, ruleID string) ([]domain.RuleLimitReview, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, scope, period, COALESCE(max_points, 0), COALESCE(max_basis_amount_minor, 0), status
		FROM paisa.rule_limits
		WHERE rule_id = $1
		ORDER BY created_at`, ruleID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	limits := []domain.RuleLimitReview{}
	for rows.Next() {
		var limit domain.RuleLimitReview
		if err := rows.Scan(&limit.ID, &limit.Scope, &limit.Period, &limit.MaxPoints, &limit.MaxBasisAmountMinor, &limit.Status); err != nil {
			return nil, AppErrorFromDB(err)
		}
		limits = append(limits, limit)
	}
	return limits, AppErrorFromDB(rows.Err())
}

func (s RuleStore) ruleDependencies(ctx context.Context, ruleID string) ([]domain.RuleDependencyReview, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT rd.id, rd.depends_on_rule_id, er.rule_key, rd.dependency_type
		FROM paisa.rule_dependencies rd
		JOIN paisa.earning_rules er ON er.id = rd.depends_on_rule_id
		WHERE rd.rule_id = $1
		ORDER BY rd.created_at`, ruleID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	deps := []domain.RuleDependencyReview{}
	for rows.Next() {
		var dep domain.RuleDependencyReview
		if err := rows.Scan(&dep.ID, &dep.DependsOnRuleID, &dep.DependsOnRuleKey, &dep.DependencyType); err != nil {
			return nil, AppErrorFromDB(err)
		}
		deps = append(deps, dep)
	}
	return deps, AppErrorFromDB(rows.Err())
}
