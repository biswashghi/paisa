package repository

import "accts-api/ports"

var _ ports.PartnerStore = PartnerStore{}
var _ ports.AuthStore = AuthStore{}
var _ ports.ProgramStore = ProgramStore{}
var _ ports.MemberStore = MemberStore{}
var _ ports.TransactionStore = TransactionStore{}
var _ ports.RuleStore = RuleStore{}
var _ ports.RewardCalculationStore = RewardCalculationStore{}
var _ ports.LedgerStore = LedgerStore{}
var _ ports.ReportingStore = ReportingStore{}
var _ ports.LocationStore = LocationStore{}
var _ ports.CatalogStore = CatalogStore{}
var _ ports.RedemptionStore = RedemptionStore{}
var _ ports.IntegrationStore = IntegrationStore{}
var _ ports.CampaignStore = CampaignStore{}
