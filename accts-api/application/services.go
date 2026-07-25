package application

import (
	"accts-api/ports"
)

type Dependencies struct {
	Stores     ports.StoreSet
	UnitOfWork ports.UnitOfWork
}

type Services struct {
	Partners         ports.PartnerService
	Programs         ports.ProgramService
	Members          ports.MemberService
	Transactions     ports.TransactionIngestionService
	RewardProcessing ports.RewardProcessingService
	Ledger           ports.LedgerService
	Reporting        ports.ReportingService
}

type app struct {
	stores ports.StoreSet
	uow    ports.UnitOfWork
}

func NewServices(deps Dependencies) Services {
	a := app{stores: deps.Stores, uow: deps.UnitOfWork}
	return Services{
		Partners:         PartnerService{app: a},
		Programs:         ProgramService{app: a},
		Members:          MemberService{app: a},
		Transactions:     TransactionIngestionService{app: a},
		RewardProcessing: RewardProcessingService{app: a},
		Ledger:           LedgerService{app: a},
		Reporting:        ReportingService{app: a},
	}
}
