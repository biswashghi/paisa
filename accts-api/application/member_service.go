package application

import (
	"context"

	"accts-api/domain"
	"accts-api/ports"
)

type MemberService struct {
	app app
}

func (s MemberService) CreateMember(ctx context.Context, partnerKey string, body domain.MemberRequest) (domain.MemberCreateResult, error) {
	var result domain.MemberCreateResult
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}
		member, err := stores.Members.Create(ctx, partner.ID, body.ExternalCustomerID)
		if err != nil {
			return err
		}
		account, err := stores.Members.CreateAccount(ctx, partner.ID, member.ID)
		if err != nil {
			return err
		}
		if err := stores.Ledger.CreateBalanceSnapshot(ctx, account.ID, partner.ID); err != nil {
			return err
		}
		for _, identifier := range body.Identifiers {
			if identifier.Type == "" || identifier.Value == "" {
				continue
			}
			valueHash := domain.HashIdentifierValue(identifier.Value)
			if err := stores.Members.InsertIdentifierHash(ctx, partner.ID, member.ID, identifier.Type, valueHash); err != nil {
				return err
			}
		}
		if body.ProgramID != "" {
			if err := enrollMember(ctx, stores, partner.ID, member.ID, body.ProgramID); err != nil {
				return err
			}
		}
		result = domain.MemberCreateResult{Member: member, Account: account}
		return nil
	})
	return result, err
}

func (s MemberService) ListMembers(ctx context.Context, partnerKey string) ([]domain.Member, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return nil, err
	}
	return s.app.stores.Members.List(ctx, partner.ID)
}

func (s MemberService) GetMember(ctx context.Context, partnerKey, memberID string) (domain.Member, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.Member{}, err
	}
	return s.app.stores.Members.GetByID(ctx, partner.ID, memberID)
}

func (s MemberService) GetRewardsProfile(ctx context.Context, partnerKey, memberID string) (domain.RewardsProfile, error) {
	partner, err := s.app.stores.Partners.GetByKey(ctx, partnerKey)
	if err != nil {
		return domain.RewardsProfile{}, err
	}
	member, err := s.app.stores.Members.GetByID(ctx, partner.ID, memberID)
	if err != nil {
		return domain.RewardsProfile{}, err
	}
	enrollment, err := s.app.stores.Members.ActiveEnrollment(ctx, partner.ID, memberID)
	if err != nil {
		return domain.RewardsProfile{}, err
	}
	addOns, err := s.app.stores.Members.ActiveRuleAssignments(ctx, partner.ID, memberID)
	if err != nil {
		return domain.RewardsProfile{}, err
	}
	balance := domain.BalanceSnapshot{}
	if accountID, err := s.app.stores.Members.AccountID(ctx, partner.ID, memberID); err == nil {
		balance, _ = s.app.stores.Ledger.GetBalance(ctx, accountID)
	}
	transactions, _ := s.app.stores.Transactions.List(ctx, partner.ID)
	memberTransactions := []domain.TransactionEvent{}
	for _, event := range transactions {
		if event.MemberID == memberID {
			memberTransactions = append(memberTransactions, event)
		}
	}
	return domain.RewardsProfile{
		Member:       member,
		Enrollment:   enrollment,
		AddOns:       addOns,
		Balance:      balance,
		Transactions: memberTransactions,
	}, nil
}

func (s MemberService) UpdateEnrollment(ctx context.Context, partnerKey, memberID string, body domain.EnrollmentRequest) error {
	return s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}
		member, err := stores.Members.GetByID(ctx, partner.ID, memberID)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("member", member.Status); err != nil {
			return err
		}
		return enrollMember(ctx, stores, partner.ID, memberID, body.ProgramID, body)
	})
}

func (s MemberService) CreateRuleAssignment(ctx context.Context, partnerKey, memberID string, body domain.MemberRuleAssignmentRequest) (domain.MemberRuleAssignment, error) {
	var assignment domain.MemberRuleAssignment
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}
		member, err := stores.Members.GetByID(ctx, partner.ID, memberID)
		if err != nil {
			return err
		}
		if err := domain.EnsureActiveStatus("member", member.Status); err != nil {
			return err
		}
		assignment, err = stores.Members.CreateRuleAssignment(ctx, partner.ID, memberID, body)
		return err
	})
	return assignment, err
}

func (s MemberService) UpdateRuleAssignment(ctx context.Context, partnerKey, memberID, assignmentID string, body domain.MemberRuleAssignmentUpdateRequest) (domain.MemberRuleAssignment, error) {
	var assignment domain.MemberRuleAssignment
	err := s.app.uow.WithinTx(ctx, func(ctx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(ctx, partnerKey)
		if err != nil {
			return err
		}
		assignment, err = stores.Members.UpdateRuleAssignment(ctx, partner.ID, memberID, assignmentID, body)
		return err
	})
	return assignment, err
}

func enrollMember(ctx context.Context, stores ports.StoreSet, partnerID, memberID, programID string, requests ...domain.EnrollmentRequest) error {
	if err := stores.Programs.EnsurePartner(ctx, partnerID, programID); err != nil {
		return err
	}
	if err := stores.Members.EndActiveEnrollment(ctx, partnerID, memberID); err != nil {
		return err
	}
	body := domain.EnrollmentRequest{}
	if len(requests) > 0 {
		body = requests[0]
	}
	_, err := stores.Members.CreateEnrollment(ctx, partnerID, memberID, programID, body)
	return err
}
