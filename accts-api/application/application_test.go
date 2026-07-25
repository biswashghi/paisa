package application

import (
	"context"
	"testing"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

func TestIngestTransactionReplaysSamePayload(t *testing.T) {
	fake := newFakeAppStore()
	body := domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  domain.EventPurchase,
		Currency:              "USD",
		EligibleMinor:         10000,
	}
	fake.transactionIdentity = ports.TransactionIdentity{ID: "evt_existing", PayloadHash: domain.HashTransactionPayload(body)}
	fake.transactionEvent = domain.TransactionEvent{ID: "evt_existing"}

	event, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt_existing" {
		t.Fatalf("expected replayed event, got %q", event.ID)
	}
	if fake.transactionCreates != 0 {
		t.Fatalf("expected no new transaction create, got %d", fake.transactionCreates)
	}
}

func TestIngestTransactionRejectsChangedIdempotencyPayload(t *testing.T) {
	fake := newFakeAppStore()
	fake.transactionIdentity = ports.TransactionIdentity{ID: "evt_existing", PayloadHash: "different"}

	_, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		EligibleMinor:         10000,
	})
	if !domain.IsErrorKind(err, domain.ErrorKindConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestIngestTransactionRejectsUnsupportedType(t *testing.T) {
	fake := newFakeAppStore()

	_, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  "exchange",
		EligibleMinor:         10000,
	})
	if !domain.IsErrorKind(err, domain.ErrorKindInvalid) {
		t.Fatalf("expected invalid type error, got %v", err)
	}
}

func TestIngestTransactionRejectsSuspendedMember(t *testing.T) {
	fake := newFakeAppStore()
	fake.member.Status = domain.StatusSuspended

	_, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  domain.EventPurchase,
		EligibleMinor:         10000,
	})
	if !domain.IsErrorKind(err, domain.ErrorKindInvariant) {
		t.Fatalf("expected invariant error, got %v", err)
	}
}

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

func TestProcessPurchaseCalculatesRewardsAndPostsLedger(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.ruleGraph = domain.RuleGraph{
		Groups: []domain.RuleGraphGroup{{ID: "group_1", Strategy: "stack"}},
		Rules: []domain.RuleGraphRule{{
			ID:       "rule_1",
			GroupID:  "group_1",
			RuleType: "points_per_dollar",
			Formula:  domain.JSONMap{"pointsPerDollar": 2},
		}},
	}

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if len(fake.createdCalculations) != 1 || fake.createdCalculations[0].PointsDelta != 200 {
		t.Fatalf("expected 200 point calculation, got %+v", fake.createdCalculations)
	}
	if len(fake.ledgerEntries) != 1 || fake.ledgerEntries[0].EntryType != domain.EntryEarn || fake.ledgerEntries[0].AvailableDelta != 200 {
		t.Fatalf("unexpected ledger entries: %+v", fake.ledgerEntries)
	}
	if fake.balance.AvailablePoints != 200 {
		t.Fatalf("expected available balance 200, got %d", fake.balance.AvailablePoints)
	}
}

func TestProcessPurchaseStacksAssignedMemberAddOnRules(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.addOnRuleVersionIDs = []string{"addon_1"}
	fake.ruleGraphs = map[string]domain.RuleGraph{
		"version_1": {
			Groups: []domain.RuleGraphGroup{{ID: "base_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "base_rule", GroupID: "base_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 1}}},
		},
		"addon_1": {
			Groups: []domain.RuleGraphGroup{{ID: "addon_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "addon_rule", GroupID: "addon_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 2}}},
		},
	}

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if got := fake.createdCalculations[0].PointsDelta; got != 300 {
		t.Fatalf("expected base 100 + add-on 200, got %d", got)
	}
	data := fake.createdCalculations[0].CalculationData
	ids, ok := data["ruleVersionIds"].([]string)
	if !ok || len(ids) != 2 || ids[0] != "version_1" || ids[1] != "addon_1" {
		t.Fatalf("expected calculation trace to include base and add-on rule versions, got %#v", data["ruleVersionIds"])
	}
}

func TestProcessPurchaseWithoutAssignmentUsesBaseRulesOnly(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_purchase"}
	fake.processingEvents["evt_purchase"] = domain.RewardProcessingEvent{
		ID:            "evt_purchase",
		PartnerID:     "partner_1",
		MemberID:      "member_1",
		Type:          domain.EventPurchase,
		EligibleMinor: 10000,
		OccurredAt:    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.ruleGraphs = map[string]domain.RuleGraph{
		"version_1": {
			Groups: []domain.RuleGraphGroup{{ID: "base_group", Strategy: domain.RuleStrategyStack}},
			Rules:  []domain.RuleGraphRule{{ID: "base_rule", GroupID: "base_group", RuleType: domain.RuleTypePointsPerDollar, Formula: domain.JSONMap{"pointsPerDollar": 1}}},
		},
	}

	if _, err := (RewardProcessingService{app: testApp(fake)}).ProcessTransactionEvents(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := fake.createdCalculations[0].PointsDelta; got != 100 {
		t.Fatalf("expected only base 100 points, got %d", got)
	}
}

func TestProcessRefundReversesOriginalCalculation(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_refund"}
	fake.processingEvents["evt_refund"] = domain.RewardProcessingEvent{
		ID:                            "evt_refund",
		PartnerID:                     "partner_1",
		MemberID:                      "member_1",
		OriginalExternalTransactionID: "txn_purchase",
		Type:                          domain.EventRefund,
		EligibleMinor:                 5000,
		OccurredAt:                    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.originalEvent = domain.RewardProcessingEvent{ID: "evt_purchase", EligibleMinor: 10000}
	fake.originalCalculation = ports.OriginalCalculation{
		ProgramID:     "program_1",
		RuleVersionID: "version_1",
		CalculationData: domain.JSONMap{"selectedAwards": []interface{}{map[string]interface{}{
			"ruleID":           "rule_1",
			"ruleVersionID":    "version_1",
			"basisAmountMinor": float64(10000),
			"points":           float64(100),
		}}},
	}
	fake.originalErr = nil
	fake.balance.AvailablePoints = 100

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if len(fake.createdCalculations) != 1 || fake.createdCalculations[0].PointsDelta != -50 {
		t.Fatalf("expected -50 point refund calculation, got %+v", fake.createdCalculations)
	}
	if fake.balance.AvailablePoints != 50 {
		t.Fatalf("expected available balance 50, got %d", fake.balance.AvailablePoints)
	}
}

func TestProcessRefundCapsCumulativeRefundBasis(t *testing.T) {
	fake := newFakeAppStore()
	fake.acceptedIDs = []string{"evt_refund"}
	fake.processingEvents["evt_refund"] = domain.RewardProcessingEvent{
		ID:                            "evt_refund",
		PartnerID:                     "partner_1",
		MemberID:                      "member_1",
		OriginalExternalTransactionID: "txn_purchase",
		Type:                          domain.EventRefund,
		EligibleMinor:                 5000,
		OccurredAt:                    time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	fake.originalEvent = domain.RewardProcessingEvent{ID: "evt_purchase", EligibleMinor: 10000}
	fake.originalCalculation = ports.OriginalCalculation{
		ProgramID:     "program_1",
		RuleVersionID: "version_1",
		CalculationData: domain.JSONMap{"selectedAwards": []interface{}{map[string]interface{}{
			"ruleID":           "rule_1",
			"ruleVersionID":    "version_1",
			"basisAmountMinor": float64(10000),
			"points":           float64(100),
		}}},
	}
	fake.originalErr = nil
	fake.priorRefundedBasis = 8000
	fake.balance.AvailablePoints = 100

	result, err := RewardProcessingService{app: testApp(fake)}.ProcessTransactionEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected processing result: %+v", result)
	}
	if len(fake.createdCalculations) != 1 || fake.createdCalculations[0].PointsDelta != -20 || fake.createdCalculations[0].BasisAmountMinor != 2000 {
		t.Fatalf("expected capped -20 point refund calculation, got %+v", fake.createdCalculations)
	}
	if fake.balance.AvailablePoints != 80 {
		t.Fatalf("expected available balance 80, got %d", fake.balance.AvailablePoints)
	}
}

func TestLedgerReserveGuardPreventsNegativeAvailable(t *testing.T) {
	fake := newFakeAppStore()
	fake.balance.AvailablePoints = 10

	_, err := postLedgerEntry(context.Background(), fake.storeSet(), ports.LedgerEntryInput{
		PartnerID:       "partner_1",
		MemberAccountID: "account_1",
		EntryType:       domain.EntryRedemptionReserve,
		AvailableDelta:  -20,
		ReservedDelta:   20,
		SourceType:      "redemption",
		SourceID:        "00000000-0000-0000-0000-000000000001",
	})
	if !domain.IsErrorKind(err, domain.ErrorKindInvariant) {
		t.Fatalf("expected invariant error, got %v", err)
	}
}

func TestGenerateLedgerLiabilityExportUsesReportingStore(t *testing.T) {
	fake := newFakeAppStore()
	fake.summary = domain.JSONMap{"liabilityPoints": 123}
	fake.exportID = "export_1"

	export, err := ReportingService{app: testApp(fake)}.GenerateLedgerLiabilityExport(context.Background(), domain.ExportRequest{
		PartnerKey:   "acme-demo",
		BusinessDate: "2026-07-25",
	})
	if err != nil {
		t.Fatal(err)
	}
	if export.ID != "export_1" || export.Summary["liabilityPoints"] != 123 {
		t.Fatalf("unexpected export: %+v", export)
	}
}

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
	ruleVersionStatus    string
	ruleVersion          domain.RuleVersion
	ruleGraph            domain.RuleGraph
	ruleGraphs           map[string]domain.RuleGraph
	publishedRuleVersion bool
	activeProgramID      string
	activeRuleVersionID  string
	addOnRuleVersionIDs  []string
	createdCalculations  []ports.RewardCalculationCreateInput
	originalEvent        domain.RewardProcessingEvent
	originalCalculation  ports.OriginalCalculation
	originalErr          error
	priorRefundedBasis   int
	ledgerEntries        []ports.LedgerEntryInput
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
	return app{stores: fake.storeSet(), uow: fake}
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

func (s fakeMemberStore) AccountID(ctx context.Context, partnerID, memberID string) (string, error) {
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

func (s fakeTransactionStore) ClaimAccepted(ctx context.Context, eventID string) error {
	if _, ok := s.state.processingEvents[eventID]; !ok {
		return domain.InvariantError("event not accepted")
	}
	return nil
}

func (s fakeTransactionStore) MarkProcessed(ctx context.Context, eventID string) error {
	return nil
}

func (s fakeTransactionStore) MarkFailed(ctx context.Context, eventID string) error {
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
