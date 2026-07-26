package application

import (
	"sync"

	"accts-api/domain"
	"accts-api/ports"
)

type Dependencies struct {
	Stores     ports.StoreSet
	UnitOfWork ports.UnitOfWork
}

type Services struct {
	Auth             ports.AuthService
	Partners         ports.PartnerService
	Programs         ports.ProgramService
	Members          ports.MemberService
	Transactions     ports.TransactionIngestionService
	RewardProcessing ports.RewardProcessingService
	Ledger           ports.LedgerService
	Reporting        ports.ReportingService
	Locations        ports.LocationService
	Catalog          ports.CatalogService
	Redemptions      ports.RedemptionService
	Integrations     ports.IntegrationService
	Dashboard        ports.DashboardService
	Campaigns        ports.CampaignService
}

type app struct {
	stores         ports.StoreSet
	uow            ports.UnitOfWork
	ruleGraphCache *ruleGraphCache
}

func NewServices(deps Dependencies) Services {
	a := app{stores: deps.Stores, uow: deps.UnitOfWork, ruleGraphCache: newRuleGraphCache()}
	return Services{
		Auth:             AuthService{app: a},
		Partners:         PartnerService{app: a},
		Programs:         ProgramService{app: a},
		Members:          MemberService{app: a},
		Transactions:     TransactionIngestionService{app: a},
		RewardProcessing: RewardProcessingService{app: a},
		Ledger:           LedgerService{app: a},
		Reporting:        ReportingService{app: a},
		Locations:        LocationService{app: a},
		Catalog:          CatalogService{app: a},
		Redemptions:      RedemptionService{app: a},
		Integrations:     IntegrationService{app: a},
		Dashboard:        DashboardService{app: a},
		Campaigns:        CampaignService{app: a},
	}
}

type ruleGraphCache struct {
	mu     sync.RWMutex
	graphs map[string]domain.RuleGraph
}

func newRuleGraphCache() *ruleGraphCache {
	return &ruleGraphCache{graphs: map[string]domain.RuleGraph{}}
}

func (c *ruleGraphCache) getMany(ids []string) (map[string]domain.RuleGraph, []string) {
	if c == nil {
		return map[string]domain.RuleGraph{}, ids
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	found := map[string]domain.RuleGraph{}
	missing := []string{}
	for _, id := range ids {
		if graph, ok := c.graphs[id]; ok {
			found[id] = graph
			continue
		}
		missing = append(missing, id)
	}
	return found, missing
}

func (c *ruleGraphCache) putMany(graphs map[string]domain.RuleGraph) {
	if c == nil || len(graphs) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, graph := range graphs {
		c.graphs[id] = graph
	}
}
