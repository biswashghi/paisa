package repository

import "accts-api/ports"

var _ ports.PartnerStore = PartnerStore{}
var _ ports.ProgramStore = ProgramStore{}
var _ ports.MemberStore = MemberStore{}
var _ ports.TransactionStore = TransactionStore{}
var _ ports.RuleStore = RuleStore{}
var _ ports.RewardCalculationStore = RewardCalculationStore{}
var _ ports.LedgerStore = LedgerStore{}
var _ ports.ReportingStore = ReportingStore{}
