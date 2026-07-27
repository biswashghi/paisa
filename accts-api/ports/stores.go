package ports

import (
	"context"
	"time"

	"accts-api/domain"
)

type StoreSet struct {
	Auth               AuthStore
	Partners           PartnerStore
	Programs           ProgramStore
	Members            MemberStore
	Transactions       TransactionStore
	Rules              RuleStore
	RewardCalculations RewardCalculationStore
	Ledger             LedgerStore
	Reporting          ReportingStore
	Locations          LocationStore
	Catalog            CatalogStore
	Redemptions        RedemptionStore
	Integrations       IntegrationStore
	Campaigns          CampaignStore
}

type UnitOfWork interface {
	WithinTx(context.Context, func(context.Context, StoreSet) error) error
}

type PartnerStore interface {
	Create(context.Context, domain.PartnerRequest) (domain.Partner, error)
	List(context.Context) ([]domain.Partner, error)
	GetByKey(context.Context, string) (domain.Partner, error)
	GetByID(context.Context, string) (domain.Partner, error)
}

type AuthStore interface {
	UpsertPartnerUserWithPassword(context.Context, string, string, string, string) (domain.PartnerUser, error)
	PartnerUserByEmail(context.Context, string) (domain.Partner, domain.PartnerUser, error)
	CreateSession(context.Context, string, string, string, time.Time) error
	RevokeSessionHash(context.Context, string) error
	AuthBySessionHash(context.Context, string) (domain.AuthContext, error)
	CreateAPIKey(context.Context, string, string, string, string, []string) (domain.APIKey, error)
	AuthByAPIKeyHash(context.Context, string) (domain.AuthContext, error)
	TouchAPIKey(context.Context, string) error
	ListAPIKeys(context.Context, string) ([]domain.APIKey, error)
	RevokeAPIKey(context.Context, string, string) error
}

type ProgramStore interface {
	Create(context.Context, string, domain.ProgramRequest) (domain.Program, error)
	List(context.Context, string) ([]domain.Program, error)
	EnsurePartner(context.Context, string, string) error
	NextRuleVersionNumber(context.Context, string, string) (int, error)
	CreateRuleVersion(context.Context, string, string, int, domain.RuleVersionRequest) (domain.RuleVersion, error)
	GetRuleVersion(context.Context, string, string, string) (domain.RuleVersion, error)
	LockRuleVersionStatus(context.Context, string, string, string) (string, error)
	ArchivePublishedRuleVersions(context.Context, string, string, string) error
	PublishRuleVersion(context.Context, string) (domain.RuleVersion, error)
	ListRuleVersions(context.Context, string, string) ([]domain.RuleVersion, error)
	ListRulePackages(context.Context, string, string) ([]domain.RuleVersion, error)
}

type MemberStore interface {
	Create(context.Context, string, string) (domain.Member, error)
	CreateAccount(context.Context, string, string) (domain.MemberAccount, error)
	InsertIdentifierHash(context.Context, string, string, string, string) error
	List(context.Context, string) ([]domain.Member, error)
	GetByID(context.Context, string, string) (domain.Member, error)
	GetByExternalID(context.Context, string, string) (domain.Member, error)
	GetByIdentifierHash(context.Context, string, string, string) (domain.Member, error)
	AccountID(context.Context, string, string) (string, error)
	EndActiveEnrollment(context.Context, string, string) error
	CreateEnrollment(context.Context, string, string, string, domain.EnrollmentRequest) (domain.ProgramEnrollment, error)
	ActiveEnrollment(context.Context, string, string) (domain.ProgramEnrollment, error)
	ListEnrollments(context.Context, string, string) ([]domain.ProgramEnrollment, error)
	CreateRuleAssignment(context.Context, string, string, domain.MemberRuleAssignmentRequest) (domain.MemberRuleAssignment, error)
	UpdateRuleAssignment(context.Context, string, string, string, domain.MemberRuleAssignmentUpdateRequest) (domain.MemberRuleAssignment, error)
	ActiveRuleAssignments(context.Context, string, string) ([]domain.MemberRuleAssignment, error)
}

type TransactionStore interface {
	FindByExternalID(context.Context, string, string) (TransactionIdentity, error)
	Create(context.Context, TransactionCreateInput) (string, error)
	CreateIfNotExists(context.Context, TransactionCreateInput) (string, bool, error)
	InsertLineItems(context.Context, string, []domain.LineItemInput) error
	Get(context.Context, string, string) (domain.TransactionEvent, error)
	List(context.Context, string) ([]domain.TransactionEvent, error)
	AcceptedIDs(context.Context, int) ([]string, error)
	ClaimAndLoadAcceptedTransactions(context.Context, int) ([]ProcessingEventBundle, error)
	ClaimAccepted(context.Context, string) error
	MarkProcessed(context.Context, string) error
	MarkFailed(context.Context, string) error
	LoadForProcessing(context.Context, string) (domain.RewardProcessingEvent, error)
	PriorProcessedPurchaseCount(context.Context, string, string, string) (int, error)
	PriorProcessedPurchaseEligibleMinorSum(context.Context, string, string, string, time.Time) (int, error)
}

type RuleStore interface {
	InsertGroups(context.Context, string, string, []domain.RuleGroupRequest) error
	ReviewGroups(context.Context, string) ([]domain.RuleGroupReview, error)
	LoadGraph(context.Context, string) (domain.RuleGraph, error)
	LoadGraphs(context.Context, []string) (map[string]domain.RuleGraph, error)
	LimitsForRule(context.Context, string) ([]domain.RuleGraphLimit, error)
	CurrentLimitUsage(context.Context, string, domain.RuleGraphLimit, time.Time) (int, int, error)
	CommitLimitUsage(context.Context, string, string, time.Time, RuleLimitUsageDelta) error
	ActiveProgramAndPublishedRules(context.Context, string, string) (string, string, error)
	ActiveProgramAndPublishedRuleSet(context.Context, string, string) (RuleSetSelection, error)
}

type RuleSetSelection struct {
	ProgramID           string
	BaseRuleVersionID   string
	AddOnRuleVersionIDs []string
	RuleVersionIDs      []string
}

type RewardCalculationStore interface {
	Get(context.Context, string, string) (domain.RewardCalculation, error)
	CreateSucceeded(context.Context, RewardCalculationCreateInput) (string, error)
	CreateFailed(context.Context, domain.RewardProcessingEvent, string) error
	OriginalForRefund(context.Context, domain.RewardProcessingEvent) (domain.RewardProcessingEvent, OriginalCalculation, error)
	PriorRefundedBasisForOriginal(context.Context, domain.RewardProcessingEvent) (int, error)
}

type LedgerStore interface {
	CreateBalanceSnapshot(context.Context, string, string) error
	GetBalance(context.Context, string) (domain.BalanceSnapshot, error)
	GetBalanceByMember(context.Context, string, string) (domain.BalanceSnapshot, error)
	LockBalance(context.Context, string) (domain.BalanceSnapshot, error)
	ListEntries(context.Context, string) ([]domain.LedgerEntry, error)
	InsertEntry(context.Context, LedgerEntryInput) (string, error)
	UpdateBalance(context.Context, domain.BalanceSnapshot) error
	PostLedgerEntry(context.Context, LedgerEntryInput) (LedgerPostResult, error)
}

type ReportingStore interface {
	LedgerSummary(context.Context, string, string) (domain.JSONMap, error)
	UpsertLedgerLiabilityExport(context.Context, string, string, domain.JSONMap) (string, error)
	ListLedgerLiabilityExports(context.Context, string) ([]domain.LedgerExport, error)
}

type TransactionIdentity struct {
	ID          string
	PayloadHash string
}

type ProcessingEventBundle struct {
	Event     domain.RewardProcessingEvent
	LineItems []domain.LineItemInput
}

type TransactionCreateInput struct {
	PartnerID                     string
	MemberID                      string
	ExternalTransactionID         string
	OriginalExternalTransactionID string
	Type                          string
	Currency                      string
	SubtotalMinor                 int
	TaxMinor                      int
	TotalMinor                    int
	EligibleMinor                 int
	OccurredAt                    time.Time
	RawPayload                    []byte
	PayloadHash                   string
	SourceSystem                  string
	SourceConnectionID            string
	SourceLocationID              string
	ExternalEventType             string
	IdempotencyKey                string
}

type LocationStore interface {
	Create(context.Context, string, domain.LocationRequest) (domain.PartnerLocation, error)
	List(context.Context, string) ([]domain.PartnerLocation, error)
}

type CatalogStore interface {
	Create(context.Context, string, domain.CatalogItemRequest) (domain.CatalogItem, error)
	List(context.Context, string) ([]domain.CatalogItem, error)
	Update(context.Context, string, string, domain.CatalogItemRequest) (domain.CatalogItem, error)
	Get(context.Context, string, string) (domain.CatalogItem, error)
	AvailableForMember(context.Context, string, string) ([]domain.CatalogItem, error)
}

type RedemptionStore interface {
	Create(context.Context, string, string, string, string, int, string, time.Time) (domain.Redemption, error)
	Get(context.Context, string, string) (domain.Redemption, error)
	GetByCode(context.Context, string, string) (domain.Redemption, error)
	UpdateStatus(context.Context, string, string, string, string) (domain.Redemption, error)
	InsertEvent(context.Context, string, string, domain.JSONMap) error
	List(context.Context, string) ([]domain.Redemption, error)
	Counts(context.Context, string) (int, error)
}

type IntegrationStore interface {
	Create(context.Context, string, domain.IntegrationConnectionRequest) (domain.IntegrationConnection, error)
	List(context.Context, string) ([]domain.IntegrationConnection, error)
	MarkSynced(context.Context, string, string) (domain.IntegrationConnection, error)
	WarningCount(context.Context, string) (int, error)
}

type CampaignStore interface {
	Create(context.Context, string, domain.CampaignRequest) (domain.Campaign, error)
	List(context.Context, string) ([]domain.Campaign, error)
}

type RuleLimitUsageDelta struct {
	LimitID     string
	UsagePoints int
	UsageBasis  int
}

type RewardCalculationCreateInput struct {
	PartnerID          string
	TransactionEventID string
	ProgramID          string
	RuleVersionID      string
	PointsDelta        int
	BasisAmountMinor   int
	CalculationData    domain.JSONMap
}

type OriginalCalculation struct {
	ID              string
	ProgramID       string
	RuleVersionID   string
	CalculationData domain.JSONMap
}

type LedgerEntryInput struct {
	PartnerID       string
	MemberAccountID string
	ProgramID       string
	EntryType       string
	AvailableDelta  int
	ReservedDelta   int
	ExpiredDelta    int
	SourceType      string
	SourceID        string
	Reason          string
	CreatedByType   string
	CreatedByID     string
}

type LedgerPostResult struct {
	LedgerEntryID string
	Balance       domain.BalanceSnapshot
}
