package repository

import (
	"context"
	"fmt"

	"accts-api/domain"
)

func EnsureProgramPartner(ctx context.Context, q Queryer, partnerID, programID string) error {
	var id string
	return q.QueryRowContext(ctx, `SELECT id FROM paisa.programs WHERE id = $1 AND partner_id = $2 AND status = 'active'`, programID, partnerID).Scan(&id)
}

func (s ProgramStore) Create(ctx context.Context, partnerID string, body domain.ProgramRequest) (domain.Program, error) {
	var program domain.Program
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.programs (partner_id, name, tier_code, priority)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		RETURNING id, partner_id, name, COALESCE(tier_code, ''), status, priority, created_at, updated_at`,
		partnerID, body.Name, body.TierCode, body.Priority,
	).Scan(&program.ID, &program.PartnerID, &program.Name, &program.TierCode, &program.Status, &program.Priority, &program.CreatedAt, &program.UpdatedAt)
	return program, AppErrorFromDB(err)
}

func (s ProgramStore) List(ctx context.Context, partnerID string) ([]domain.Program, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, name, COALESCE(tier_code, ''), status, priority, created_at, updated_at
		FROM paisa.programs
		WHERE partner_id = $1
		ORDER BY priority, created_at`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	programs := []domain.Program{}
	for rows.Next() {
		var program domain.Program
		if err := rows.Scan(&program.ID, &program.PartnerID, &program.Name, &program.TierCode, &program.Status, &program.Priority, &program.CreatedAt, &program.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		programs = append(programs, program)
	}
	return programs, AppErrorFromDB(rows.Err())
}

func (s ProgramStore) EnsurePartner(ctx context.Context, partnerID, programID string) error {
	return AppErrorFromDB(EnsureProgramPartner(ctx, s.q, partnerID, programID))
}

func (s ProgramStore) NextRuleVersionNumber(ctx context.Context, programID, scope string) (int, error) {
	versionNumber := 1
	err := s.q.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM paisa.program_rule_versions WHERE program_id = $1 AND scope = $2`, programID, scope).Scan(&versionNumber)
	return versionNumber, AppErrorFromDB(err)
}

func (s ProgramStore) CreateRuleVersion(ctx context.Context, partnerID, programID string, versionNumber int, body domain.RuleVersionRequest) (domain.RuleVersion, error) {
	var version domain.RuleVersion
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.program_rule_versions (partner_id, program_id, version, scope, rule_set_key, name, description, earn_basis)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8)
		RETURNING id, partner_id, program_id, version, status, scope, COALESCE(rule_set_key, ''), COALESCE(name, ''), COALESCE(description, ''), earn_basis, published_at, created_at, updated_at`,
		partnerID, programID, versionNumber, body.Scope, body.RuleSetKey, body.Name, body.Description, body.EarnBasis,
	).Scan(&version.ID, &version.PartnerID, &version.ProgramID, &version.Version, &version.Status, &version.Scope, &version.RuleSetKey, &version.Name, &version.Description, &version.EarnBasis, &version.PublishedAt, &version.CreatedAt, &version.UpdatedAt)
	return version, AppErrorFromDB(err)
}

func (s ProgramStore) GetRuleVersion(ctx context.Context, partnerID, programID, versionID string) (domain.RuleVersion, error) {
	var version domain.RuleVersion
	err := s.q.QueryRowContext(ctx, `
		SELECT id, partner_id, program_id, version, status, scope, COALESCE(rule_set_key, ''), COALESCE(name, ''), COALESCE(description, ''), earn_basis, published_at, created_at, updated_at
		FROM paisa.program_rule_versions
		WHERE id = $1 AND partner_id = $2 AND program_id = $3`, versionID, partnerID, programID,
	).Scan(&version.ID, &version.PartnerID, &version.ProgramID, &version.Version, &version.Status, &version.Scope, &version.RuleSetKey, &version.Name, &version.Description, &version.EarnBasis, &version.PublishedAt, &version.CreatedAt, &version.UpdatedAt)
	return version, AppErrorFromDB(err)
}

func (s ProgramStore) LockRuleVersionStatus(ctx context.Context, partnerID, programID, versionID string) (string, error) {
	var currentStatus string
	err := s.q.QueryRowContext(ctx, `
		SELECT status
		FROM paisa.program_rule_versions
		WHERE id = $1 AND program_id = $2 AND partner_id = $3
		FOR UPDATE`, versionID, programID, partnerID).Scan(&currentStatus)
	return currentStatus, AppErrorFromDB(err)
}

func (s ProgramStore) ArchivePublishedRuleVersions(ctx context.Context, partnerID, programID, scope string) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.program_rule_versions
		SET status = $1, updated_at = now()
		WHERE partner_id = $2 AND program_id = $3 AND scope = $4 AND status = $5`, domain.RuleArchived, partnerID, programID, scope, domain.RulePublished)
	return AppErrorFromDB(err)
}

func (s ProgramStore) PublishRuleVersion(ctx context.Context, versionID string) (domain.RuleVersion, error) {
	var version domain.RuleVersion
	err := s.q.QueryRowContext(ctx, `
		UPDATE paisa.program_rule_versions
		SET status = $1, published_at = now(), updated_at = now()
		WHERE id = $2
		RETURNING id, partner_id, program_id, version, status, scope, COALESCE(rule_set_key, ''), COALESCE(name, ''), COALESCE(description, ''), earn_basis, published_at, created_at, updated_at`,
		domain.RulePublished, versionID,
	).Scan(&version.ID, &version.PartnerID, &version.ProgramID, &version.Version, &version.Status, &version.Scope, &version.RuleSetKey, &version.Name, &version.Description, &version.EarnBasis, &version.PublishedAt, &version.CreatedAt, &version.UpdatedAt)
	return version, AppErrorFromDB(err)
}

func (s ProgramStore) ListRuleVersions(ctx context.Context, partnerID, programID string) ([]domain.RuleVersion, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, program_id, version, status, scope, COALESCE(rule_set_key, ''), COALESCE(name, ''), COALESCE(description, ''), earn_basis, published_at, created_at, updated_at
		FROM paisa.program_rule_versions
		WHERE partner_id = $1 AND program_id = $2
		ORDER BY version DESC`, partnerID, programID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	versions := []domain.RuleVersion{}
	for rows.Next() {
		var version domain.RuleVersion
		if err := rows.Scan(&version.ID, &version.PartnerID, &version.ProgramID, &version.Version, &version.Status, &version.Scope, &version.RuleSetKey, &version.Name, &version.Description, &version.EarnBasis, &version.PublishedAt, &version.CreatedAt, &version.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		versions = append(versions, version)
	}
	return versions, AppErrorFromDB(rows.Err())
}

func (s ProgramStore) ListRulePackages(ctx context.Context, partnerID, programID string) ([]domain.RuleVersion, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, program_id, version, status, scope, COALESCE(rule_set_key, ''), COALESCE(name, ''), COALESCE(description, ''), earn_basis, published_at, created_at, updated_at
		FROM paisa.program_rule_versions
		WHERE partner_id = $1 AND program_id = $2 AND scope = $3
		ORDER BY status, version DESC`, partnerID, programID, domain.RuleScopeMemberAddOn)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	versions := []domain.RuleVersion{}
	for rows.Next() {
		var version domain.RuleVersion
		if err := rows.Scan(&version.ID, &version.PartnerID, &version.ProgramID, &version.Version, &version.Status, &version.Scope, &version.RuleSetKey, &version.Name, &version.Description, &version.EarnBasis, &version.PublishedAt, &version.CreatedAt, &version.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		versions = append(versions, version)
	}
	return versions, AppErrorFromDB(rows.Err())
}

func (s RuleStore) InsertGroups(ctx context.Context, partnerID, versionID string, groups []domain.RuleGroupRequest) error {
	for groupIndex, group := range groups {
		group.Priority = domain.DefaultPriority(group.Priority, groupIndex)
		group.ResolutionStrategy = normalizeDefault(group.ResolutionStrategy, "stack")
		var groupID string
		if err := s.q.QueryRowContext(ctx, `
			INSERT INTO paisa.rule_groups (partner_id, rule_version_id, name, priority, resolution_strategy)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`, partnerID, versionID, group.Name, group.Priority, group.ResolutionStrategy,
		).Scan(&groupID); err != nil {
			return AppErrorFromDB(err)
		}

		ruleIDsByKey := map[string]string{}
		pendingDependencies := []struct {
			ruleID     string
			dependency domain.RuleDependencyRequest
		}{}

		for ruleIndex, rule := range group.Rules {
			rule.Priority = domain.DefaultPriority(rule.Priority, ruleIndex)
			rule.Status = normalizeDefault(rule.Status, "active")
			rule.RuleKey = normalizeDefault(rule.RuleKey, fmt.Sprintf("rule-%d", ruleIndex+1))
			eligibilityJSON := mustJSON(domain.DefaultMap(rule.EligibilityConfig))
			formulaJSON := mustJSON(domain.DefaultMap(rule.FormulaConfig))

			var ruleID string
			if err := s.q.QueryRowContext(ctx, `
				INSERT INTO paisa.earning_rules (
					partner_id, rule_group_id, rule_key, name, rule_type, priority, status,
					eligibility_config, formula_config
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id`,
				partnerID, groupID, rule.RuleKey, rule.Name, normalizeDefault(rule.RuleType, "points_per_dollar"),
				rule.Priority, rule.Status, eligibilityJSON, formulaJSON,
			).Scan(&ruleID); err != nil {
				return AppErrorFromDB(err)
			}
			ruleIDsByKey[rule.RuleKey] = ruleID

			for _, limit := range rule.Limits {
				if _, err := s.q.ExecContext(ctx, `
					INSERT INTO paisa.rule_limits (
						partner_id, rule_id, scope, period, max_points, max_basis_amount_minor
					)
					VALUES ($1, $2, $3, $4, NULLIF($5, 0), NULLIF($6, 0))`,
					partnerID, ruleID, normalizeDefault(limit.Scope, "member"), normalizeDefault(limit.Period, "lifetime"),
					limit.MaxPoints, limit.MaxBasisAmountMinor,
				); err != nil {
					return AppErrorFromDB(err)
				}
			}
			for _, dependency := range rule.Dependencies {
				pendingDependencies = append(pendingDependencies, struct {
					ruleID     string
					dependency domain.RuleDependencyRequest
				}{ruleID: ruleID, dependency: dependency})
			}
		}

		for _, pending := range pendingDependencies {
			dependsOnRuleID := ruleIDsByKey[pending.dependency.DependsOnRuleKey]
			if dependsOnRuleID == "" {
				return domain.InvalidError(fmt.Sprintf("unknown dependsOnRuleKey %q", pending.dependency.DependsOnRuleKey))
			}
			if _, err := s.q.ExecContext(ctx, `
				INSERT INTO paisa.rule_dependencies (partner_id, rule_id, depends_on_rule_id, dependency_type)
				VALUES ($1, $2, $3, $4)`,
				partnerID, pending.ruleID, dependsOnRuleID, normalizeDefault(pending.dependency.DependencyType, "requires_match"),
			); err != nil {
				return AppErrorFromDB(err)
			}
		}
	}
	return nil
}

func ValidateRuleVersionForPublish(ctx context.Context, q Queryer, versionID string) error {
	graph, err := RuleStore{q: q}.LoadGraph(ctx, versionID)
	if err != nil {
		return err
	}
	return domain.ValidateRuleGraphForPublish(graph.Groups, graph.Rules, graph.Dependencies, graph.Limits)
}
