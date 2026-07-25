package domain

import "time"

type Partner struct {
	ID         string    `json:"id"`
	PartnerKey string    `json:"partnerKey"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Program struct {
	ID        string    `json:"id"`
	PartnerID string    `json:"partnerId"`
	Name      string    `json:"name"`
	TierCode  string    `json:"tierCode,omitempty"`
	Status    string    `json:"status"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type RuleVersion struct {
	ID          string     `json:"id"`
	PartnerID   string     `json:"partnerId"`
	ProgramID   string     `json:"programId"`
	Version     int        `json:"version"`
	Status      string     `json:"status"`
	Scope       string     `json:"scope"`
	RuleSetKey  string     `json:"ruleSetKey,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	EarnBasis   string     `json:"earnBasis"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Member struct {
	ID                 string    `json:"id"`
	PartnerID          string    `json:"partnerId"`
	ExternalCustomerID string    `json:"externalCustomerId"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type MemberAccount struct {
	ID        string `json:"id"`
	PartnerID string `json:"partnerId"`
	MemberID  string `json:"memberId"`
	Status    string `json:"status"`
}

type BalanceSnapshot struct {
	MemberAccountID string    `json:"memberAccountId"`
	PartnerID       string    `json:"partnerId"`
	AvailablePoints int       `json:"availablePoints"`
	ReservedPoints  int       `json:"reservedPoints"`
	ExpiredPoints   int       `json:"expiredPoints"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type TransactionEvent struct {
	ID                            string          `json:"id"`
	PartnerID                     string          `json:"partnerId"`
	MemberID                      string          `json:"memberId"`
	ExternalTransactionID         string          `json:"externalTransactionId"`
	OriginalExternalTransactionID string          `json:"originalExternalTransactionId,omitempty"`
	Type                          string          `json:"type"`
	Status                        string          `json:"status"`
	Currency                      string          `json:"currency"`
	SubtotalMinor                 int             `json:"subtotalMinor"`
	TaxMinor                      int             `json:"taxMinor"`
	TotalMinor                    int             `json:"totalMinor"`
	EligibleMinor                 int             `json:"eligibleMinor"`
	OccurredAt                    time.Time       `json:"occurredAt"`
	RawPayload                    JSONMap         `json:"-"`
	LineItems                     []LineItemInput `json:"lineItems,omitempty"`
	CreatedAt                     time.Time       `json:"createdAt"`
	UpdatedAt                     time.Time       `json:"updatedAt"`
}

type RewardProcessingEvent struct {
	ID                            string
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
	RawPayload                    JSONMap
	Lines                         []LineItemInput
}

type RewardCalculation struct {
	ID                 string    `json:"id"`
	PartnerID          string    `json:"partnerId"`
	TransactionEventID string    `json:"transactionEventId"`
	ProgramID          string    `json:"programId,omitempty"`
	RuleVersionID      string    `json:"ruleVersionId,omitempty"`
	Status             string    `json:"status"`
	PointsDelta        int       `json:"pointsDelta"`
	BasisAmountMinor   int       `json:"basisAmountMinor"`
	CalculationData    JSONMap   `json:"calculationData"`
	FailureReason      string    `json:"failureReason,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}

type LedgerEntry struct {
	ID              string    `json:"id"`
	PartnerID       string    `json:"partnerId"`
	MemberAccountID string    `json:"memberAccountId"`
	ProgramID       string    `json:"programId,omitempty"`
	EntryType       string    `json:"entryType"`
	AvailableDelta  int       `json:"availableDelta"`
	ReservedDelta   int       `json:"reservedDelta"`
	ExpiredDelta    int       `json:"expiredDelta"`
	SourceType      string    `json:"sourceType"`
	SourceID        string    `json:"sourceId"`
	Reason          string    `json:"reason,omitempty"`
	CreatedByType   string    `json:"createdByType"`
	CreatedByID     string    `json:"createdById,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

type PartnerRequest struct {
	PartnerKey string `json:"partnerKey"`
	Name       string `json:"name"`
}

type ProgramRequest struct {
	Name     string `json:"name"`
	TierCode string `json:"tierCode"`
	Priority int    `json:"priority"`
}

type RuleVersionRequest struct {
	EarnBasis   string             `json:"earnBasis"`
	Scope       string             `json:"scope,omitempty"`
	RuleSetKey  string             `json:"ruleSetKey,omitempty"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	RuleGroups  []RuleGroupRequest `json:"ruleGroups"`
}

type RuleVersionReview struct {
	RuleVersion RuleVersion          `json:"ruleVersion"`
	Groups      []RuleGroupReview    `json:"groups"`
	Validation  RuleValidationResult `json:"validation"`
}

type RuleValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

type RuleGroupReview struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	ResolutionStrategy string              `json:"resolutionStrategy"`
	Priority           int                 `json:"priority"`
	Status             string              `json:"status"`
	Rules              []EarningRuleReview `json:"rules"`
}

type EarningRuleReview struct {
	ID                string                 `json:"id"`
	RuleKey           string                 `json:"ruleKey"`
	Name              string                 `json:"name"`
	RuleType          string                 `json:"ruleType"`
	Priority          int                    `json:"priority"`
	Status            string                 `json:"status"`
	EligibilityConfig JSONMap                `json:"eligibilityConfig"`
	FormulaConfig     JSONMap                `json:"formulaConfig"`
	Limits            []RuleLimitReview      `json:"limits"`
	Dependencies      []RuleDependencyReview `json:"dependencies"`
}

type RuleLimitReview struct {
	ID                  string `json:"id"`
	Scope               string `json:"scope"`
	Period              string `json:"period"`
	MaxPoints           int    `json:"maxPoints,omitempty"`
	MaxBasisAmountMinor int    `json:"maxBasisAmountMinor,omitempty"`
	Status              string `json:"status"`
}

type RuleDependencyReview struct {
	ID               string `json:"id"`
	DependsOnRuleID  string `json:"dependsOnRuleId"`
	DependsOnRuleKey string `json:"dependsOnRuleKey"`
	DependencyType   string `json:"dependencyType"`
}

type RuleGroupRequest struct {
	Name               string        `json:"name"`
	ResolutionStrategy string        `json:"resolutionStrategy"`
	Priority           int           `json:"priority"`
	Rules              []RuleRequest `json:"rules"`
}

type RuleRequest struct {
	RuleKey           string                  `json:"ruleKey"`
	Name              string                  `json:"name"`
	RuleType          string                  `json:"ruleType"`
	Priority          int                     `json:"priority"`
	Status            string                  `json:"status"`
	EligibilityConfig map[string]interface{}  `json:"eligibilityConfig"`
	FormulaConfig     map[string]interface{}  `json:"formulaConfig"`
	Limits            []RuleLimitRequest      `json:"limits"`
	Dependencies      []RuleDependencyRequest `json:"dependencies"`
}

type RuleLimitRequest struct {
	Scope               string `json:"scope"`
	Period              string `json:"period"`
	MaxPoints           int    `json:"maxPoints"`
	MaxBasisAmountMinor int    `json:"maxBasisAmountMinor"`
}

type RuleDependencyRequest struct {
	DependsOnRuleKey string `json:"dependsOnRuleKey"`
	DependencyType   string `json:"dependencyType"`
}

type MemberCreateResult struct {
	Member  Member        `json:"member"`
	Account MemberAccount `json:"account"`
}

type MemberRequest struct {
	ExternalCustomerID string              `json:"externalCustomerId"`
	Identifiers        []IdentifierRequest `json:"identifiers"`
	ProgramID          string              `json:"programId"`
}

type IdentifierRequest struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type EnrollmentRequest struct {
	ProgramID     string `json:"programId"`
	EffectiveAt   string `json:"effectiveAt,omitempty"`
	ChangeReason  string `json:"changeReason,omitempty"`
	CreatedByType string `json:"createdByType,omitempty"`
	CreatedByID   string `json:"createdById,omitempty"`
}

type ProgramEnrollment struct {
	ID            string     `json:"id"`
	PartnerID     string     `json:"partnerId"`
	MemberID      string     `json:"memberId"`
	ProgramID     string     `json:"programId"`
	ProgramName   string     `json:"programName,omitempty"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"startedAt"`
	EndedAt       *time.Time `json:"endedAt,omitempty"`
	EffectiveAt   *time.Time `json:"effectiveAt,omitempty"`
	ChangeReason  string     `json:"changeReason,omitempty"`
	CreatedByType string     `json:"createdByType"`
	CreatedByID   string     `json:"createdById,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type MemberRuleAssignment struct {
	ID                  string     `json:"id"`
	PartnerID           string     `json:"partnerId"`
	MemberID            string     `json:"memberId"`
	ProgramEnrollmentID string     `json:"programEnrollmentId"`
	RuleVersionID       string     `json:"ruleVersionId"`
	RuleSetKey          string     `json:"ruleSetKey,omitempty"`
	Name                string     `json:"name,omitempty"`
	Description         string     `json:"description,omitempty"`
	ProgramID           string     `json:"programId,omitempty"`
	Status              string     `json:"status"`
	StartsAt            time.Time  `json:"startsAt"`
	EndsAt              *time.Time `json:"endsAt,omitempty"`
	Reason              string     `json:"reason,omitempty"`
	CreatedByType       string     `json:"createdByType"`
	CreatedByID         string     `json:"createdById,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type MemberRuleAssignmentRequest struct {
	RuleVersionID string `json:"ruleVersionId"`
	StartsAt      string `json:"startsAt,omitempty"`
	Reason        string `json:"reason,omitempty"`
	CreatedByType string `json:"createdByType,omitempty"`
	CreatedByID   string `json:"createdById,omitempty"`
}

type MemberRuleAssignmentUpdateRequest struct {
	Status string `json:"status"`
	EndsAt string `json:"endsAt,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type RewardsProfile struct {
	Member       Member                 `json:"member"`
	Enrollment   ProgramEnrollment      `json:"enrollment"`
	AddOns       []MemberRuleAssignment `json:"addOns"`
	Balance      BalanceSnapshot        `json:"balance"`
	Transactions []TransactionEvent     `json:"transactions"`
}

type TransactionIngestRequest struct {
	ExternalTransactionID         string          `json:"externalTransactionId"`
	ExternalCustomerID            string          `json:"externalCustomerId"`
	OriginalExternalTransactionID string          `json:"originalExternalTransactionId"`
	Type                          string          `json:"type"`
	Currency                      string          `json:"currency"`
	SubtotalMinor                 int             `json:"subtotalMinor"`
	TaxMinor                      int             `json:"taxMinor"`
	TotalMinor                    int             `json:"totalMinor"`
	EligibleMinor                 int             `json:"eligibleMinor"`
	OccurredAt                    string          `json:"occurredAt"`
	LineItems                     []LineItemInput `json:"lineItems"`
}

type LineItemInput struct {
	ExternalLineID string `json:"externalLineId"`
	SKU            string `json:"sku"`
	Category       string `json:"category"`
	Quantity       int    `json:"quantity"`
	SubtotalMinor  int    `json:"subtotalMinor"`
	TaxMinor       int    `json:"taxMinor"`
	TotalMinor     int    `json:"totalMinor"`
	EligibleMinor  int    `json:"eligibleMinor"`
}

type AdjustmentRequest struct {
	AvailableDelta int    `json:"availableDelta"`
	ReservedDelta  int    `json:"reservedDelta"`
	ExpiredDelta   int    `json:"expiredDelta"`
	Reason         string `json:"reason"`
	SourceID       string `json:"sourceId"`
}

type AdjustmentResult struct {
	LedgerEntryID string `json:"ledgerEntryId"`
}

type ProcessTransactionEventsResult struct {
	Processed int `json:"processed"`
	Failed    int `json:"failed"`
}

type ExportRequest struct {
	PartnerKey   string `json:"partnerKey"`
	BusinessDate string `json:"businessDate"`
}

type LedgerExport struct {
	ID           string    `json:"id"`
	PartnerID    string    `json:"partnerId"`
	BusinessDate string    `json:"businessDate"`
	Status       string    `json:"status"`
	FilePath     string    `json:"filePath"`
	Summary      JSONMap   `json:"summary"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type JSONMap map[string]interface{}
