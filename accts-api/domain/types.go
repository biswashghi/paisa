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

type AuthContext struct {
	PartnerID     string    `json:"partnerId"`
	PartnerKey    string    `json:"partnerKey"`
	PartnerName   string    `json:"partnerName"`
	ActorType     string    `json:"actorType"`
	ActorID       string    `json:"actorId"`
	Scopes        []string  `json:"scopes"`
	Authenticated time.Time `json:"authenticatedAt"`
}

type PartnerUser struct {
	ID           string    `json:"id"`
	PartnerID    string    `json:"partnerId"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token   string      `json:"token"`
	Auth    AuthContext `json:"auth"`
	Partner Partner     `json:"partner"`
	User    PartnerUser `json:"user"`
}

type PartnerOnboardRequest struct {
	PartnerKey    string `json:"partnerKey"`
	PartnerName   string `json:"partnerName"`
	AdminEmail    string `json:"adminEmail"`
	AdminName     string `json:"adminName"`
	AdminPassword string `json:"adminPassword"`
}

type PartnerOnboardResult struct {
	Partner Partner     `json:"partner"`
	User    PartnerUser `json:"user"`
}

type APIKey struct {
	ID         string     `json:"id"`
	PartnerID  string     `json:"partnerId"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"keyPrefix"`
	Scopes     []string   `json:"scopes"`
	Status     string     `json:"status"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

type APIKeyCreateRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type APIKeyCreateResult struct {
	APIKey APIKey `json:"apiKey"`
	Token  string `json:"token"`
}

type PartnerLocation struct {
	ID                 string    `json:"id"`
	PartnerID          string    `json:"partnerId"`
	Name               string    `json:"name"`
	Address            string    `json:"address,omitempty"`
	Timezone           string    `json:"timezone"`
	Status             string    `json:"status"`
	ExternalLocationID string    `json:"externalLocationId,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type LocationRequest struct {
	Name               string `json:"name"`
	Address            string `json:"address"`
	Timezone           string `json:"timezone"`
	ExternalLocationID string `json:"externalLocationId"`
}

type IntegrationConnection struct {
	ID                 string     `json:"id"`
	PartnerID          string     `json:"partnerId"`
	Provider           string     `json:"provider"`
	Status             string     `json:"status"`
	ExternalMerchantID string     `json:"externalMerchantId,omitempty"`
	ExternalLocationID string     `json:"externalLocationId,omitempty"`
	Metadata           JSONMap    `json:"metadata"`
	LastSyncAt         *time.Time `json:"lastSyncAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type IntegrationConnectionRequest struct {
	Provider           string  `json:"provider"`
	Status             string  `json:"status"`
	ExternalMerchantID string  `json:"externalMerchantId"`
	ExternalLocationID string  `json:"externalLocationId"`
	Metadata           JSONMap `json:"metadata"`
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

type ResolveMemberRequest struct {
	ExternalCustomerID string              `json:"externalCustomerId"`
	Identifiers        []IdentifierRequest `json:"identifiers"`
	ProgramID          string              `json:"programId"`
}

type ResolveMemberResult struct {
	Member  Member        `json:"member"`
	Account MemberAccount `json:"account"`
	Created bool          `json:"created"`
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

type NormalizedTransaction struct {
	SourceSystem                  string               `json:"sourceSystem"`
	SourceConnectionID            string               `json:"sourceConnectionId,omitempty"`
	SourceLocationID              string               `json:"sourceLocationId,omitempty"`
	ExternalEventType             string               `json:"externalEventType,omitempty"`
	IdempotencyKey                string               `json:"idempotencyKey,omitempty"`
	ExternalTransactionID         string               `json:"externalTransactionId"`
	OriginalExternalTransactionID string               `json:"originalExternalTransactionId,omitempty"`
	Customer                      ResolveMemberRequest `json:"customer"`
	MemberID                      string               `json:"memberId,omitempty"`
	Type                          string               `json:"type"`
	Currency                      string               `json:"currency"`
	SubtotalMinor                 int                  `json:"subtotalMinor"`
	TaxMinor                      int                  `json:"taxMinor"`
	TotalMinor                    int                  `json:"totalMinor"`
	EligibleMinor                 int                  `json:"eligibleMinor"`
	OccurredAt                    string               `json:"occurredAt"`
	LineItems                     []LineItemInput      `json:"lineItems"`
	RawPayload                    JSONMap              `json:"rawPayload"`
}

type ManualTransactionRequest struct {
	MemberID              string               `json:"memberId"`
	Customer              ResolveMemberRequest `json:"customer"`
	ExternalTransactionID string               `json:"externalTransactionId"`
	LocationID            string               `json:"locationId"`
	Category              string               `json:"category"`
	Currency              string               `json:"currency"`
	SubtotalMinor         int                  `json:"subtotalMinor"`
	TaxMinor              int                  `json:"taxMinor"`
	TotalMinor            int                  `json:"totalMinor"`
	EligibleMinor         int                  `json:"eligibleMinor"`
	OccurredAt            string               `json:"occurredAt"`
	LineItems             []LineItemInput      `json:"lineItems"`
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

type CatalogItem struct {
	ID                  string    `json:"id"`
	PartnerID           string    `json:"partnerId"`
	ProgramID           string    `json:"programId,omitempty"`
	LocationID          string    `json:"locationId,omitempty"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty"`
	PointsCost          int       `json:"pointsCost"`
	RewardType          string    `json:"rewardType"`
	Status              string    `json:"status"`
	ExpiresAfterMinutes int       `json:"expiresAfterMinutes"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type CatalogItemRequest struct {
	ProgramID           string `json:"programId"`
	LocationID          string `json:"locationId"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	PointsCost          int    `json:"pointsCost"`
	RewardType          string `json:"rewardType"`
	Status              string `json:"status"`
	ExpiresAfterMinutes int    `json:"expiresAfterMinutes"`
}

type Redemption struct {
	ID                   string     `json:"id"`
	PartnerID            string     `json:"partnerId"`
	MemberID             string     `json:"memberId"`
	MemberAccountID      string     `json:"memberAccountId"`
	CatalogItemID        string     `json:"catalogItemId"`
	CatalogItemName      string     `json:"catalogItemName,omitempty"`
	Code                 string     `json:"code"`
	Status               string     `json:"status"`
	PointsCost           int        `json:"pointsCost"`
	ReservationExpiresAt *time.Time `json:"reservationExpiresAt,omitempty"`
	FailureReason        string     `json:"failureReason,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type RedemptionRequest struct {
	MemberID      string `json:"memberId"`
	CatalogItemID string `json:"catalogItemId"`
}

type RedemptionActionResult struct {
	Redemption    Redemption      `json:"redemption"`
	LedgerEntryID string          `json:"ledgerEntryId,omitempty"`
	Balance       BalanceSnapshot `json:"balance"`
}

type DashboardSummary struct {
	Partner             Partner `json:"partner"`
	ActiveLocations     int     `json:"activeLocations"`
	ActiveCatalogItems  int     `json:"activeCatalogItems"`
	OpenRedemptions     int     `json:"openRedemptions"`
	IntegrationWarnings int     `json:"integrationWarnings"`
}

type Campaign struct {
	ID                  string     `json:"id"`
	PartnerID           string     `json:"partnerId"`
	Name                string     `json:"name"`
	Description         string     `json:"description,omitempty"`
	Status              string     `json:"status"`
	StartsAt            *time.Time `json:"startsAt,omitempty"`
	EndsAt              *time.Time `json:"endsAt,omitempty"`
	RequiredVisitCount  int        `json:"requiredVisitCount"`
	RewardCatalogItemID string     `json:"rewardCatalogItemId,omitempty"`
	Metadata            JSONMap    `json:"metadata"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type CampaignRequest struct {
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	Status              string  `json:"status"`
	StartsAt            string  `json:"startsAt"`
	EndsAt              string  `json:"endsAt"`
	RequiredVisitCount  int     `json:"requiredVisitCount"`
	RewardCatalogItemID string  `json:"rewardCatalogItemId"`
	Metadata            JSONMap `json:"metadata"`
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
