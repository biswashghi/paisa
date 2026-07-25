package application

import "accts-api/ports"

var _ ports.PartnerService = PartnerService{}
var _ ports.ProgramService = ProgramService{}
var _ ports.MemberService = MemberService{}
var _ ports.TransactionIngestionService = TransactionIngestionService{}
var _ ports.RewardProcessingService = RewardProcessingService{}
var _ ports.LedgerService = LedgerService{}
var _ ports.ReportingService = ReportingService{}
