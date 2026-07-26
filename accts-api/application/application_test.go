package application

import (
	"context"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

type fakeAppStore struct {
	*fakeState
}

type fakeState struct {
	partner              domain.Partner
	member               domain.Member
	account              domain.MemberAccount
	accountID            string
	balance              domain.BalanceSnapshot
	transactionIdentity  ports.TransactionIdentity
	transactionEvent     domain.TransactionEvent
	transactionCreates   int
	acceptedIDs          []string
	processingEvents     map[string]domain.RewardProcessingEvent
	markedFailed         int
	markedProcessed      int
	ruleVersionStatus    string
	ruleVersion          domain.RuleVersion
	ruleGraph            domain.RuleGraph
	ruleGraphs           map[string]domain.RuleGraph
	loadGraphsCalls      int
	publishedRuleVersion bool
	activeProgramID      string
	activeRuleVersionID  string
	addOnRuleVersionIDs  []string
	activeRuleSetErr     error
	createdCalculations  []ports.RewardCalculationCreateInput
	failedCalculations   int
	originalEvent        domain.RewardProcessingEvent
	originalCalculation  ports.OriginalCalculation
	originalErr          error
	priorRefundedBasis   int
	ledgerEntries        []ports.LedgerEntryInput
	accountIDCalls       int
	balanceByMemberCalls int
	summary              domain.JSONMap
	exportID             string
}

type fakePartnerStore struct{ state *fakeState }
type fakeProgramStore struct{ state *fakeState }
type fakeMemberStore struct{ state *fakeState }
type fakeTransactionStore struct{ state *fakeState }
type fakeRuleStore struct{ state *fakeState }
type fakeRewardCalculationStore struct{ state *fakeState }
type fakeLedgerStore struct{ state *fakeState }
type fakeReportingStore struct{ state *fakeState }

func newFakeAppStore() *fakeAppStore {
	state := &fakeState{
		partner:             domain.Partner{ID: "partner_1", PartnerKey: "acme-demo", Status: domain.StatusActive},
		member:              domain.Member{ID: "member_1", PartnerID: "partner_1", ExternalCustomerID: "cust_1", Status: domain.StatusActive},
		account:             domain.MemberAccount{ID: "account_1", PartnerID: "partner_1", MemberID: "member_1", Status: domain.StatusActive},
		accountID:           "account_1",
		balance:             domain.BalanceSnapshot{MemberAccountID: "account_1", PartnerID: "partner_1"},
		processingEvents:    map[string]domain.RewardProcessingEvent{},
		ruleVersionStatus:   domain.RuleDraft,
		ruleVersion:         domain.RuleVersion{ID: "version_1", PartnerID: "partner_1", ProgramID: "program_1", Scope: domain.RuleScopeProgramBase},
		activeProgramID:     "program_1",
		activeRuleVersionID: "version_1",
		originalErr:         notFoundErr(),
		summary:             domain.JSONMap{},
		exportID:            "export_1",
	}
	return &fakeAppStore{fakeState: state}
}

func testApp(fake *fakeAppStore) app {
	return app{stores: fake.storeSet(), uow: fake, ruleGraphCache: newRuleGraphCache()}
}

func (f *fakeAppStore) storeSet() ports.StoreSet {
	return ports.StoreSet{
		Partners:           fakePartnerStore{state: f.fakeState},
		Programs:           fakeProgramStore{state: f.fakeState},
		Members:            fakeMemberStore{state: f.fakeState},
		Transactions:       fakeTransactionStore{state: f.fakeState},
		Rules:              fakeRuleStore{state: f.fakeState},
		RewardCalculations: fakeRewardCalculationStore{state: f.fakeState},
		Ledger:             fakeLedgerStore{state: f.fakeState},
		Reporting:          fakeReportingStore{state: f.fakeState},
	}
}

func (f *fakeAppStore) WithinTx(ctx context.Context, fn func(context.Context, ports.StoreSet) error) error {
	return fn(ctx, f.storeSet())
}

func notFoundErr() error {
	return domain.AppError{Kind: domain.ErrorKindNotFound, Message: "not found"}
}

func (s fakePartnerStore) Create(ctx context.Context, body domain.PartnerRequest) (domain.Partner, error) {
	s.state.partner = domain.Partner{ID: "partner_1", PartnerKey: body.PartnerKey, Name: body.Name, Status: domain.StatusActive}
	return s.state.partner, nil
}

func (s fakePartnerStore) List(ctx context.Context) ([]domain.Partner, error) {
	return []domain.Partner{s.state.partner}, nil
}

func (s fakePartnerStore) GetByKey(ctx context.Context, partnerKey string) (domain.Partner, error) {
	if partnerKey != s.state.partner.PartnerKey {
		return domain.Partner{}, notFoundErr()
	}
	return s.state.partner, nil
}

func (s fakePartnerStore) GetByID(ctx context.Context, partnerID string) (domain.Partner, error) {
	if partnerID != s.state.partner.ID {
		return domain.Partner{}, notFoundErr()
	}
	return s.state.partner, nil
}

func (s fakeProgramStore) Create(ctx context.Context, partnerID string, body domain.ProgramRequest) (domain.Program, error) {
	return domain.Program{ID: "program_1", PartnerID: partnerID, Name: body.Name, Status: domain.StatusActive}, nil
}

func (s fakeProgramStore) List(ctx context.Context, partnerID string) ([]domain.Program, error) {
	return []domain.Program{{ID: "program_1", PartnerID: partnerID, Status: domain.StatusActive}}, nil
}

func (s fakeProgramStore) EnsurePartner(ctx context.Context, partnerID, programID string) error {
	return nil
}

func (s fakeProgramStore) NextRuleVersionNumber(ctx context.Context, programID, scope string) (int, error) {
	return 1, nil
}

func (s fakeProgramStore) CreateRuleVersion(ctx context.Context, partnerID, programID string, versionNumber int, body domain.RuleVersionRequest) (domain.RuleVersion, error) {
	s.state.ruleVersion = domain.RuleVersion{ID: "version_1", PartnerID: partnerID, ProgramID: programID, Version: versionNumber, Scope: body.Scope, RuleSetKey: body.RuleSetKey, Name: body.Name, Description: body.Description, EarnBasis: body.EarnBasis}
	return s.state.ruleVersion, nil
}

func (s fakeProgramStore) GetRuleVersion(ctx context.Context, partnerID, programID, versionID string) (domain.RuleVersion, error) {
	return s.state.ruleVersion, nil
}

func (s fakeProgramStore) LockRuleVersionStatus(ctx context.Context, partnerID, programID, versionID string) (string, error) {
	return s.state.ruleVersionStatus, nil
}

func (s fakeProgramStore) ArchivePublishedRuleVersions(ctx context.Context, partnerID, programID, scope string) error {
	return nil
}

func (s fakeProgramStore) PublishRuleVersion(ctx context.Context, versionID string) (domain.RuleVersion, error) {
	s.state.publishedRuleVersion = true
	return s.state.ruleVersion, nil
}

func (s fakeProgramStore) ListRuleVersions(ctx context.Context, partnerID, programID string) ([]domain.RuleVersion, error) {
	return []domain.RuleVersion{s.state.ruleVersion}, nil
}

func (s fakeProgramStore) ListRulePackages(ctx context.Context, partnerID, programID string) ([]domain.RuleVersion, error) {
	return []domain.RuleVersion{}, nil
}

func (s fakeMemberStore) Create(ctx context.Context, partnerID, externalCustomerID string) (domain.Member, error) {
	s.state.member = domain.Member{ID: "member_1", PartnerID: partnerID, ExternalCustomerID: externalCustomerID, Status: domain.StatusActive}
	return s.state.member, nil
}

func (s fakeMemberStore) CreateAccount(ctx context.Context, partnerID, memberID string) (domain.MemberAccount, error) {
	s.state.account = domain.MemberAccount{ID: s.state.accountID, PartnerID: partnerID, MemberID: memberID, Status: domain.StatusActive}
	return s.state.account, nil
}

func (s fakeMemberStore) InsertIdentifierHash(ctx context.Context, partnerID, memberID, identifierType, valueHash string) error {
	return nil
}

func (s fakeMemberStore) List(ctx context.Context, partnerID string) ([]domain.Member, error) {
	return []domain.Member{s.state.member}, nil
}

func (s fakeMemberStore) GetByID(ctx context.Context, partnerID, memberID string) (domain.Member, error) {
	return s.state.member, nil
}

func (s fakeMemberStore) GetByExternalID(ctx context.Context, partnerID, externalCustomerID string) (domain.Member, error) {
	return s.state.member, nil
}

func (s fakeMemberStore) GetByIdentifierHash(ctx context.Context, partnerID, identifierType, valueHash string) (domain.Member, error) {
	return domain.Member{}, notFoundErr()
}

func (s fakeMemberStore) AccountID(ctx context.Context, partnerID, memberID string) (string, error) {
	s.state.accountIDCalls++
	return s.state.accountID, nil
}

func (s fakeMemberStore) EndActiveEnrollment(ctx context.Context, partnerID, memberID string) error {
	return nil
}

func (s fakeMemberStore) CreateEnrollment(ctx context.Context, partnerID, memberID, programID string, body domain.EnrollmentRequest) (domain.ProgramEnrollment, error) {
	return domain.ProgramEnrollment{ID: "enrollment_1", PartnerID: partnerID, MemberID: memberID, ProgramID: programID, Status: domain.StatusActive, CreatedByType: "system"}, nil
}

func (s fakeMemberStore) ActiveEnrollment(ctx context.Context, partnerID, memberID string) (domain.ProgramEnrollment, error) {
	return domain.ProgramEnrollment{ID: "enrollment_1", PartnerID: partnerID, MemberID: memberID, ProgramID: s.state.activeProgramID, Status: domain.StatusActive, CreatedByType: "system"}, nil
}

func (s fakeMemberStore) ListEnrollments(ctx context.Context, partnerID, memberID string) ([]domain.ProgramEnrollment, error) {
	return []domain.ProgramEnrollment{{ID: "enrollment_1", PartnerID: partnerID, MemberID: memberID, ProgramID: s.state.activeProgramID, Status: domain.StatusActive, CreatedByType: "system"}}, nil
}

func (s fakeMemberStore) CreateRuleAssignment(ctx context.Context, partnerID, memberID string, body domain.MemberRuleAssignmentRequest) (domain.MemberRuleAssignment, error) {
	return domain.MemberRuleAssignment{ID: "assignment_1", PartnerID: partnerID, MemberID: memberID, RuleVersionID: body.RuleVersionID, Status: domain.StatusActive, CreatedByType: "system"}, nil
}

func (s fakeMemberStore) UpdateRuleAssignment(ctx context.Context, partnerID, memberID, assignmentID string, body domain.MemberRuleAssignmentUpdateRequest) (domain.MemberRuleAssignment, error) {
	return domain.MemberRuleAssignment{ID: assignmentID, PartnerID: partnerID, MemberID: memberID, Status: body.Status, CreatedByType: "system"}, nil
}

func (s fakeMemberStore) ActiveRuleAssignments(ctx context.Context, partnerID, memberID string) ([]domain.MemberRuleAssignment, error) {
	return []domain.MemberRuleAssignment{}, nil
}

func (s fakeTransactionStore) FindByExternalID(ctx context.Context, partnerID, externalTransactionID string) (ports.TransactionIdentity, error) {
	if s.state.transactionIdentity.ID == "" {
		return ports.TransactionIdentity{}, notFoundErr()
	}
	return s.state.transactionIdentity, nil
}

func (s fakeTransactionStore) Create(ctx context.Context, input ports.TransactionCreateInput) (string, error) {
	s.state.transactionCreates++
	s.state.transactionEvent = domain.TransactionEvent{ID: "evt_created", PartnerID: input.PartnerID, MemberID: input.MemberID, ExternalTransactionID: input.ExternalTransactionID}
	return s.state.transactionEvent.ID, nil
}

func (s fakeTransactionStore) CreateIfNotExists(ctx context.Context, input ports.TransactionCreateInput) (string, bool, error) {
	if s.state.transactionIdentity.ID != "" {
		return "", false, nil
	}
	s.state.transactionCreates++
	s.state.transactionEvent = domain.TransactionEvent{ID: "evt_created", PartnerID: input.PartnerID, MemberID: input.MemberID, ExternalTransactionID: input.ExternalTransactionID}
	s.state.transactionIdentity = ports.TransactionIdentity{ID: s.state.transactionEvent.ID, PayloadHash: input.PayloadHash}
	return s.state.transactionEvent.ID, true, nil
}

func (s fakeTransactionStore) InsertLineItems(ctx context.Context, eventID string, lines []domain.LineItemInput) error {
	return nil
}

func (s fakeTransactionStore) Get(ctx context.Context, partnerID, eventID string) (domain.TransactionEvent, error) {
	if s.state.transactionEvent.ID == "" {
		return domain.TransactionEvent{ID: eventID, PartnerID: partnerID}, nil
	}
	return s.state.transactionEvent, nil
}

func (s fakeTransactionStore) List(ctx context.Context, partnerID string) ([]domain.TransactionEvent, error) {
	return []domain.TransactionEvent{s.state.transactionEvent}, nil
}

func (s fakeTransactionStore) AcceptedIDs(ctx context.Context, limit int) ([]string, error) {
	return s.state.acceptedIDs, nil
}

func (s fakeTransactionStore) ClaimAndLoadAcceptedTransactions(ctx context.Context, limit int) ([]ports.ProcessingEventBundle, error) {
	bundles := []ports.ProcessingEventBundle{}
	for _, id := range s.state.acceptedIDs {
		event, ok := s.state.processingEvents[id]
		if !ok {
			continue
		}
		bundles = append(bundles, ports.ProcessingEventBundle{Event: event, LineItems: event.Lines})
	}
	return bundles, nil
}

func (s fakeTransactionStore) ClaimAccepted(ctx context.Context, eventID string) error {
	if _, ok := s.state.processingEvents[eventID]; !ok {
		return domain.InvariantError("event not accepted")
	}
	return nil
}

func (s fakeTransactionStore) MarkProcessed(ctx context.Context, eventID string) error {
	s.state.markedProcessed++
	return nil
}

func (s fakeTransactionStore) MarkFailed(ctx context.Context, eventID string) error {
	s.state.markedFailed++
	return nil
}

func (s fakeTransactionStore) LoadForProcessing(ctx context.Context, eventID string) (domain.RewardProcessingEvent, error) {
	event, ok := s.state.processingEvents[eventID]
	if !ok {
		return domain.RewardProcessingEvent{}, notFoundErr()
	}
	return event, nil
}

func (s fakeTransactionStore) PriorProcessedPurchaseCount(ctx context.Context, partnerID, memberID, excludeEventID string) (int, error) {
	return 0, nil
}

func (s fakeTransactionStore) PriorProcessedPurchaseEligibleMinorSum(ctx context.Context, partnerID, memberID, excludeEventID string, since time.Time) (int, error) {
	return 0, nil
}

func (s fakeRuleStore) InsertGroups(ctx context.Context, partnerID, versionID string, groups []domain.RuleGroupRequest) error {
	return nil
}

func (s fakeRuleStore) ReviewGroups(ctx context.Context, ruleVersionID string) ([]domain.RuleGroupReview, error) {
	return nil, nil
}

func (s fakeRuleStore) LoadGraph(ctx context.Context, ruleVersionID string) (domain.RuleGraph, error) {
	if s.state.ruleGraphs != nil {
		if graph, ok := s.state.ruleGraphs[ruleVersionID]; ok {
			return graph, nil
		}
	}
	return s.state.ruleGraph, nil
}

func (s fakeRuleStore) LoadGraphs(ctx context.Context, ruleVersionIDs []string) (map[string]domain.RuleGraph, error) {
	s.state.loadGraphsCalls++
	graphs := map[string]domain.RuleGraph{}
	for _, id := range ruleVersionIDs {
		graph, err := s.LoadGraph(ctx, id)
		if err != nil {
			return nil, err
		}
		graphs[id] = graph
	}
	return graphs, nil
}

func (s fakeRuleStore) LimitsForRule(ctx context.Context, ruleID string) ([]domain.RuleGraphLimit, error) {
	return nil, nil
}

func (s fakeRuleStore) CurrentLimitUsage(ctx context.Context, memberID string, limit domain.RuleGraphLimit, occurredAt time.Time) (int, int, error) {
	return 0, 0, nil
}

func (s fakeRuleStore) CommitLimitUsage(ctx context.Context, partnerID, memberID string, occurredAt time.Time, delta ports.RuleLimitUsageDelta) error {
	return nil
}

func (s fakeRuleStore) ActiveProgramAndPublishedRules(ctx context.Context, partnerID, memberID string) (string, string, error) {
	return s.state.activeProgramID, s.state.activeRuleVersionID, nil
}

func (s fakeRuleStore) ActiveProgramAndPublishedRuleSet(ctx context.Context, partnerID, memberID string) (ports.RuleSetSelection, error) {
	if s.state.activeRuleSetErr != nil {
		return ports.RuleSetSelection{}, s.state.activeRuleSetErr
	}
	all := append([]string{s.state.activeRuleVersionID}, s.state.addOnRuleVersionIDs...)
	return ports.RuleSetSelection{
		ProgramID:           s.state.activeProgramID,
		BaseRuleVersionID:   s.state.activeRuleVersionID,
		AddOnRuleVersionIDs: s.state.addOnRuleVersionIDs,
		RuleVersionIDs:      all,
	}, nil
}

func (s fakeRewardCalculationStore) Get(ctx context.Context, partnerID, transactionEventID string) (domain.RewardCalculation, error) {
	return domain.RewardCalculation{ID: "calc_1", PartnerID: partnerID, TransactionEventID: transactionEventID}, nil
}

func (s fakeRewardCalculationStore) CreateSucceeded(ctx context.Context, input ports.RewardCalculationCreateInput) (string, error) {
	s.state.createdCalculations = append(s.state.createdCalculations, input)
	return "00000000-0000-0000-0000-000000000001", nil
}

func (s fakeRewardCalculationStore) CreateFailed(ctx context.Context, event domain.RewardProcessingEvent, reason string) error {
	s.state.failedCalculations++
	return nil
}

func (s fakeRewardCalculationStore) OriginalForRefund(ctx context.Context, event domain.RewardProcessingEvent) (domain.RewardProcessingEvent, ports.OriginalCalculation, error) {
	if s.state.originalErr != nil {
		return domain.RewardProcessingEvent{}, ports.OriginalCalculation{}, s.state.originalErr
	}
	return s.state.originalEvent, s.state.originalCalculation, nil
}

func (s fakeRewardCalculationStore) PriorRefundedBasisForOriginal(ctx context.Context, event domain.RewardProcessingEvent) (int, error) {
	return s.state.priorRefundedBasis, nil
}

func (s fakeLedgerStore) CreateBalanceSnapshot(ctx context.Context, accountID, partnerID string) error {
	s.state.balance = domain.BalanceSnapshot{MemberAccountID: accountID, PartnerID: partnerID}
	return nil
}

func (s fakeLedgerStore) GetBalance(ctx context.Context, accountID string) (domain.BalanceSnapshot, error) {
	return s.state.balance, nil
}

func (s fakeLedgerStore) GetBalanceByMember(ctx context.Context, partnerID, memberID string) (domain.BalanceSnapshot, error) {
	s.state.balanceByMemberCalls++
	return s.state.balance, nil
}

func (s fakeLedgerStore) LockBalance(ctx context.Context, accountID string) (domain.BalanceSnapshot, error) {
	return s.state.balance, nil
}

func (s fakeLedgerStore) ListEntries(ctx context.Context, accountID string) ([]domain.LedgerEntry, error) {
	return nil, nil
}

func (s fakeLedgerStore) InsertEntry(ctx context.Context, input ports.LedgerEntryInput) (string, error) {
	s.state.ledgerEntries = append(s.state.ledgerEntries, input)
	return "ledger_1", nil
}

func (s fakeLedgerStore) UpdateBalance(ctx context.Context, balance domain.BalanceSnapshot) error {
	s.state.balance = balance
	return nil
}

func (s fakeLedgerStore) PostLedgerEntry(ctx context.Context, input ports.LedgerEntryInput) (ports.LedgerPostResult, error) {
	balance, err := s.LockBalance(ctx, input.MemberAccountID)
	if err != nil {
		return ports.LedgerPostResult{}, err
	}
	next, err := domain.ApplyLedgerDelta(balance, input.EntryType, input.AvailableDelta, input.ReservedDelta, input.ExpiredDelta)
	if err != nil {
		return ports.LedgerPostResult{}, err
	}
	entryID, err := s.InsertEntry(ctx, input)
	if err != nil {
		return ports.LedgerPostResult{}, err
	}
	if err := s.UpdateBalance(ctx, next); err != nil {
		return ports.LedgerPostResult{}, err
	}
	return ports.LedgerPostResult{LedgerEntryID: entryID, Balance: next}, nil
}

func (s fakeReportingStore) LedgerSummary(ctx context.Context, partnerID, businessDate string) (domain.JSONMap, error) {
	return s.state.summary, nil
}

func (s fakeReportingStore) UpsertLedgerLiabilityExport(ctx context.Context, partnerID, businessDate string, summary domain.JSONMap) (string, error) {
	s.state.summary = summary
	return s.state.exportID, nil
}

func (s fakeReportingStore) ListLedgerLiabilityExports(ctx context.Context, partnerID string) ([]domain.LedgerExport, error) {
	return []domain.LedgerExport{{ID: s.state.exportID, PartnerID: partnerID, Summary: s.state.summary}}, nil
}
