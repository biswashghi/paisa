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
	ListMembers(context.Context, string) ([]domain.Member, error)
	GetMember(context.Context, string, string) (domain.Member, error)
	GetRewardsProfile(context.Context, string, string) (domain.RewardsProfile, error)
	UpdateEnrollment(context.Context, string, string, domain.EnrollmentRequest) error
	CreateRuleAssignment(context.Context, string, string, domain.MemberRuleAssignmentRequest) (domain.MemberRuleAssignment, error)
	UpdateRuleAssignment(context.Context, string, string, string, domain.MemberRuleAssignmentUpdateRequest) (domain.MemberRuleAssignment, error)
}

type TransactionIngestionService interface {
	IngestTransaction(context.Context, string, domain.TransactionIngestRequest) (domain.TransactionEvent, error)
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
