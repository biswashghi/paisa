package domain

const (
	EventPurchase = "purchase"
	EventRefund   = "refund"

	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusClosed    = "closed"

	StatusAccepted   = "accepted"
	StatusProcessing = "processing"
	StatusProcessed  = "processed"
	StatusFailed     = "failed"

	RuleDraft     = "draft"
	RulePublished = "published"
	RuleArchived  = "archived"

	RuleScopeProgramBase = "program_base"
	RuleScopeMemberAddOn = "member_add_on"

	EntryEarn               = "earn"
	EntryRefund             = "refund"
	EntryAdjustment         = "adjustment"
	EntryRedemptionReserve  = "redemption_reserve"
	EntryRedemptionCapture  = "redemption_capture"
	EntryReservationRelease = "reservation_release"
	EntryPointsExpiration   = "points_expiration"

	RedemptionRequested = "requested"
	RedemptionReserved  = "reserved"
	RedemptionValidated = "validated"
	RedemptionCaptured  = "captured"
	RedemptionReleased  = "released"
	RedemptionExpired   = "expired"
	RedemptionVoided    = "voided"

	RuleStrategyStack     = "stack"
	RuleStrategyMaxOf     = "max_of"
	RuleStrategyWaterfall = "waterfall"

	RuleTypePointsPerDollar     = "points_per_dollar"
	RuleTypeFixedPerTransaction = "fixed_per_transaction"
	RuleTypeFirstPurchaseBonus  = "first_purchase_bonus"
	RuleTypeSpendWindowBonus    = "spend_window_bonus"
)
