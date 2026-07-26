package application

import "accts-api/ports"

var _ ports.PartnerService = PartnerService{}
var _ ports.AuthService = AuthService{}
var _ ports.ProgramService = ProgramService{}
var _ ports.MemberService = MemberService{}
var _ ports.TransactionIngestionService = TransactionIngestionService{}
var _ ports.RewardProcessingService = RewardProcessingService{}
var _ ports.LedgerService = LedgerService{}
var _ ports.ReportingService = ReportingService{}
var _ ports.LocationService = LocationService{}
var _ ports.CatalogService = CatalogService{}
var _ ports.RedemptionService = RedemptionService{}
var _ ports.IntegrationService = IntegrationService{}
var _ ports.DashboardService = DashboardService{}
var _ ports.CampaignService = CampaignService{}
