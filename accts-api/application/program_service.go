package application

import (
	"context"

	"accts-api/domain"
	"accts-api/ports"
)

type ProgramService struct {
	app app
}

func (s ProgramService) CreateProgram(ctx context.Context, partnerKey string, body domain.ProgramRequest) (domain.Program, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.Program{}, err
	}
	if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
		return domain.Program{}, err
	}
	return s.app.stores.Programs.Create(ctx, partner.ID, body)
}

func (s ProgramService) ListPrograms(ctx context.Context, partnerKey string) ([]domain.Program, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	return s.app.stores.Programs.List(ctx, partner.ID)
}

func (s ProgramService) CreateRuleVersion(ctx context.Context, partnerKey, programID string, body domain.RuleVersionRequest) (domain.RuleVersion, error) {
	body.EarnBasis = normalizeDefault(body.EarnBasis, "eligible")
	body.Scope = normalizeDefault(body.Scope, domain.RuleScopeProgramBase)
	if body.Scope == domain.RuleScopeMemberAddOn && body.RuleSetKey == "" {
		body.RuleSetKey = body.Name
	}
	var version domain.RuleVersion
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}
		if err := stores.Programs.EnsurePartner(ctx, partner.ID, programID); err != nil {
			return err
		}
		versionNumber, err := stores.Programs.NextRuleVersionNumber(ctx, programID, body.Scope)
		if err != nil {
			return err
		}
		version, err = stores.Programs.CreateRuleVersion(ctx, partner.ID, programID, versionNumber, body)
		if err != nil {
			return err
		}
		return stores.Rules.InsertGroups(ctx, partner.ID, version.ID, body.RuleGroups)
	})
	return version, err
}

func (s ProgramService) PublishRuleVersion(ctx context.Context, partnerKey, programID, versionID string) (domain.RuleVersion, error) {
	var version domain.RuleVersion
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}
		currentStatus, err := stores.Programs.LockRuleVersionStatus(ctx, partner.ID, programID, versionID)
		if err != nil {
			return err
		}
		if currentStatus != domain.RuleDraft {
			return domain.ConflictError("only draft rule versions can be published")
		}
		graph, err := stores.Rules.LoadGraph(ctx, versionID)
		if err != nil {
			return err
		}
		if err := domain.ValidateRuleGraphForPublish(graph.Groups, graph.Rules, graph.Dependencies, graph.Limits); err != nil {
			return err
		}
		versionMeta, err := stores.Programs.GetRuleVersion(ctx, partner.ID, programID, versionID)
		if err != nil {
			return err
		}
		if versionMeta.Scope == domain.RuleScopeProgramBase {
			if err := stores.Programs.ArchivePublishedRuleVersions(ctx, partner.ID, programID, domain.RuleScopeProgramBase); err != nil {
				return err
			}
		}
		version, err = stores.Programs.PublishRuleVersion(ctx, versionID)
		return err
	})
	return version, err
}

func (s ProgramService) ListRulePackages(ctx context.Context, partnerKey, programID string) ([]domain.RuleVersion, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	if err := s.app.stores.Programs.EnsurePartner(ctx, partner.ID, programID); err != nil {
		return nil, err
	}
	return s.app.stores.Programs.ListRulePackages(ctx, partner.ID, programID)
}

func (s ProgramService) CreateRulePackage(ctx context.Context, partnerKey, programID string, body domain.RuleVersionRequest) (domain.RuleVersion, error) {
	body.Scope = domain.RuleScopeMemberAddOn
	if body.Name == "" {
		body.Name = body.RuleSetKey
	}
	if body.RuleSetKey == "" {
		body.RuleSetKey = body.Name
	}
	return s.CreateRuleVersion(ctx, partnerKey, programID, body)
}

func (s ProgramService) ListRuleVersions(ctx context.Context, partnerKey, programID string) ([]domain.RuleVersion, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	if err := s.app.stores.Programs.EnsurePartner(ctx, partner.ID, programID); err != nil {
		return nil, err
	}
	return s.app.stores.Programs.ListRuleVersions(ctx, partner.ID, programID)
}

func (s ProgramService) GetRuleVersionReview(ctx context.Context, partnerKey, programID, versionID string) (domain.RuleVersionReview, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.RuleVersionReview{}, err
	}
	if err := s.app.stores.Programs.EnsurePartner(ctx, partner.ID, programID); err != nil {
		return domain.RuleVersionReview{}, err
	}
	version, err := s.app.stores.Programs.GetRuleVersion(ctx, partner.ID, programID, versionID)
	if err != nil {
		return domain.RuleVersionReview{}, err
	}
	groups, err := s.app.stores.Rules.ReviewGroups(ctx, version.ID)
	if err != nil {
		return domain.RuleVersionReview{}, err
	}
	return domain.RuleVersionReview{
		RuleVersion: version,
		Groups:      groups,
		Validation:  s.ValidateRuleVersion(ctx, partnerKey, programID, versionID),
	}, nil
}

func (s ProgramService) ValidateRuleVersion(ctx context.Context, partnerKey, programID, versionID string) domain.RuleValidationResult {
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := stores.Programs.EnsurePartner(ctx, partner.ID, programID); err != nil {
			return err
		}
		if _, err := stores.Programs.GetRuleVersion(ctx, partner.ID, programID, versionID); err != nil {
			return err
		}
		graph, err := stores.Rules.LoadGraph(ctx, versionID)
		if err != nil {
			return err
		}
		return domain.ValidateRuleGraphForPublish(graph.Groups, graph.Rules, graph.Dependencies, graph.Limits)
	})
	if err != nil {
		return domain.RuleValidationResult{Valid: false, Errors: []string{err.Error()}}
	}
	return domain.RuleValidationResult{Valid: true, Errors: []string{}}
}
