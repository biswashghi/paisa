package ports

import (
	"context"

	"accts-api/domain"
)

type PartnerService interface {
	Create(context.Context, domain.PartnerRequest) (domain.Partner, error)
	List(context.Context) ([]domain.Partner, error)
	GetByKey(context.Context, string) (domain.Partner, error)
}

type AuthService interface {
	OnboardPartner(context.Context, domain.PartnerOnboardRequest) (domain.PartnerOnboardResult, error)
	Login(context.Context, domain.LoginRequest) (domain.LoginResult, error)
	Logout(context.Context, string) error
	AuthenticateToken(context.Context, string) (domain.AuthContext, error)
	CreateAPIKey(context.Context, domain.AuthContext, domain.APIKeyCreateRequest) (domain.APIKeyCreateResult, error)
	ListAPIKeys(context.Context, domain.AuthContext) ([]domain.APIKey, error)
	RevokeAPIKey(context.Context, domain.AuthContext, string) error
}

type ProgramService interface {
	CreateProgram(context.Context, string, domain.ProgramRequest) (domain.Program, error)
	ListPrograms(context.Context, string) ([]domain.Program, error)
	CreateRuleVersion(context.Context, string, string, domain.RuleVersionRequest) (domain.RuleVersion, error)
	CreateRulePackage(context.Context, string, string, domain.RuleVersionRequest) (domain.RuleVersion, error)
	ListRulePackages(context.Context, string, string) ([]domain.RuleVersion, error)
	PublishRuleVersion(context.Context, string, string, string) (domain.RuleVersion, error)
	ListRuleVersions(context.Context, string, string) ([]domain.RuleVersion, error)
	GetRuleVersionReview(context.Context, string, string, string) (domain.RuleVersionReview, error)
	ValidateRuleVersion(context.Context, string, string, string) domain.RuleValidationResult
}

type MemberService interface {
	CreateMember(context.Context, string, domain.MemberRequest) (domain.MemberCreateResult, error)
	ResolveOrCreateMember(context.Context, string, domain.ResolveMemberRequest) (domain.ResolveMemberResult, error)
	ResolveOrCreateMemberForPartnerID(context.Context, string, domain.ResolveMemberRequest) (domain.ResolveMemberResult, error)
	ListMembers(context.Context, string) ([]domain.Member, error)
	GetMember(context.Context, string, string) (domain.Member, error)
	GetRewardsProfile(context.Context, string, string) (domain.RewardsProfile, error)
	UpdateEnrollment(context.Context, string, string, domain.EnrollmentRequest) error
	CreateRuleAssignment(context.Context, string, string, domain.MemberRuleAssignmentRequest) (domain.MemberRuleAssignment, error)
	UpdateRuleAssignment(context.Context, string, string, string, domain.MemberRuleAssignmentUpdateRequest) (domain.MemberRuleAssignment, error)
}

type TransactionIngestionService interface {
	IngestTransaction(context.Context, string, domain.TransactionIngestRequest) (domain.TransactionEvent, error)
	IngestManualTransaction(context.Context, domain.AuthContext, domain.ManualTransactionRequest) (domain.TransactionEvent, error)
	IngestNormalizedTransaction(context.Context, string, domain.NormalizedTransaction) (domain.TransactionEvent, error)
	GetTransaction(context.Context, string, string) (domain.TransactionEvent, error)
	ListTransactions(context.Context, string) ([]domain.TransactionEvent, error)
	GetCalculation(context.Context, string, string) (domain.RewardCalculation, error)
}

type RewardProcessingService interface {
	ProcessTransactionEvents(context.Context) (domain.ProcessTransactionEventsResult, error)
}

type LedgerService interface {
	GetBalance(context.Context, string, string) (domain.BalanceSnapshot, error)
	GetLedger(context.Context, string, string) ([]domain.LedgerEntry, error)
	CreateAdjustment(context.Context, string, string, domain.AdjustmentRequest) (domain.AdjustmentResult, error)
}

type ReportingService interface {
	GenerateLedgerLiabilityExport(context.Context, domain.ExportRequest) (domain.LedgerExport, error)
	ListLedgerLiabilityExports(context.Context, string) ([]domain.LedgerExport, error)
}

type LocationService interface {
	CreateLocation(context.Context, domain.AuthContext, domain.LocationRequest) (domain.PartnerLocation, error)
	ListLocations(context.Context, domain.AuthContext) ([]domain.PartnerLocation, error)
}

type CatalogService interface {
	CreateCatalogItem(context.Context, domain.AuthContext, domain.CatalogItemRequest) (domain.CatalogItem, error)
	ListCatalogItems(context.Context, domain.AuthContext) ([]domain.CatalogItem, error)
	UpdateCatalogItem(context.Context, domain.AuthContext, string, domain.CatalogItemRequest) (domain.CatalogItem, error)
	AvailableRewards(context.Context, domain.AuthContext, string) ([]domain.CatalogItem, error)
}

type RedemptionService interface {
	CreateRedemption(context.Context, domain.AuthContext, domain.RedemptionRequest) (domain.RedemptionActionResult, error)
	ValidateRedemption(context.Context, domain.AuthContext, string) (domain.RedemptionActionResult, error)
	CaptureRedemption(context.Context, domain.AuthContext, string) (domain.RedemptionActionResult, error)
	ReleaseRedemption(context.Context, domain.AuthContext, string) (domain.RedemptionActionResult, error)
	ListRedemptions(context.Context, domain.AuthContext) ([]domain.Redemption, error)
}

type IntegrationService interface {
	ListConnections(context.Context, domain.AuthContext) ([]domain.IntegrationConnection, error)
	StartSquareOAuth(context.Context, domain.AuthContext) (domain.IntegrationConnection, error)
	CompleteSquareOAuth(context.Context, domain.AuthContext, string) (domain.IntegrationConnection, error)
	SyncConnection(context.Context, domain.AuthContext, string) (domain.IntegrationConnection, error)
}

type DashboardService interface {
	Summary(context.Context, domain.AuthContext) (domain.DashboardSummary, error)
}

type CampaignService interface {
	CreateCampaign(context.Context, domain.AuthContext, domain.CampaignRequest) (domain.Campaign, error)
	ListCampaigns(context.Context, domain.AuthContext) ([]domain.Campaign, error)
}
