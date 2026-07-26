package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"accts-api/domain"
	"accts-api/ports"

	"github.com/gorilla/mux"
)

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

type Server struct {
	services Services
}

func NewRouter(services Services) *mux.Router {
	server := Server{services: services}
	r := mux.NewRouter()
	r.HandleFunc("/health", server.health).Methods("GET")

	v1 := r.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/partners", server.createPartner).Methods("POST")
	v1.HandleFunc("/partners", server.listPartners).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}", server.getPartner).Methods("GET")

	v1.HandleFunc("/partners/{partnerKey}/programs", server.createProgram).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/programs", server.listPrograms).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-versions", server.createRuleVersion).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-versions", server.listRuleVersions).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-versions/{versionId}", server.getRuleVersionReview).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-versions/{versionId}/validate", server.validateRuleVersion).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-versions/{versionId}/publish", server.publishRuleVersion).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-packages", server.listRulePackages).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/programs/{programId}/rule-packages", server.createRulePackage).Methods("POST")

	v1.HandleFunc("/partners/{partnerKey}/members", server.createMember).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/members", server.listMembers).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}", server.getMember).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/rewards-profile", server.getRewardsProfile).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/program-enrollment", server.updateMemberEnrollment).Methods("PUT")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/rule-assignments", server.createRuleAssignment).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/rule-assignments/{assignmentId}", server.updateRuleAssignment).Methods("PATCH")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/balance", server.getMemberBalance).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/ledger", server.getMemberLedger).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/members/{memberId}/adjustments", server.createAdjustment).Methods("POST")

	v1.HandleFunc("/partners/{partnerKey}/ingest/transactions", server.ingestTransaction).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/ingest/transactions/{transactionEventId}", server.getIngestedTransaction).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/transactions", server.listTransactions).Methods("GET")
	v1.HandleFunc("/partners/{partnerKey}/transactions/{transactionEventId}/calculation", server.getTransactionCalculation).Methods("GET")

	v1.HandleFunc("/jobs/process-transaction-events", server.processTransactionEvents).Methods("POST")
	v1.HandleFunc("/jobs/generate-ledger-liability-export", server.generateLedgerLiabilityExport).Methods("POST")
	v1.HandleFunc("/partners/{partnerKey}/exports/ledger-liability", server.listLedgerLiabilityExports).Methods("GET")

	partner := r.PathPrefix("/partner/v1").Subrouter()
	partner.HandleFunc("/auth/login", server.partnerLogin).Methods("POST")
	partner.HandleFunc("/auth/logout", server.partnerLogout).Methods("POST")

	partnerAuthed := partner.NewRoute().Subrouter()
	partnerAuthed.Use(server.requireAuth)
	partnerAuthed.HandleFunc("/me", server.partnerMe).Methods("GET")
	partnerAuthed.HandleFunc("/programs", server.partnerCreateProgram).Methods("POST")
	partnerAuthed.HandleFunc("/programs", server.partnerListPrograms).Methods("GET")
	partnerAuthed.HandleFunc("/programs/{programId}/rule-versions", server.partnerCreateRuleVersion).Methods("POST")
	partnerAuthed.HandleFunc("/programs/{programId}/rule-versions", server.partnerListRuleVersions).Methods("GET")
	partnerAuthed.HandleFunc("/programs/{programId}/rule-versions/{versionId}", server.partnerGetRuleVersionReview).Methods("GET")
	partnerAuthed.HandleFunc("/programs/{programId}/rule-versions/{versionId}/publish", server.partnerPublishRuleVersion).Methods("POST")
	partnerAuthed.HandleFunc("/programs/{programId}/rule-packages", server.partnerListRulePackages).Methods("GET")
	partnerAuthed.HandleFunc("/programs/{programId}/rule-packages", server.partnerCreateRulePackage).Methods("POST")
	partnerAuthed.HandleFunc("/members", server.partnerCreateMember).Methods("POST")
	partnerAuthed.HandleFunc("/members", server.partnerListMembers).Methods("GET")
	partnerAuthed.HandleFunc("/members/{memberId}/rewards-profile", server.partnerGetRewardsProfile).Methods("GET")
	partnerAuthed.HandleFunc("/members/{memberId}/program-enrollment", server.partnerUpdateMemberEnrollment).Methods("PUT")
	partnerAuthed.HandleFunc("/members/{memberId}/rule-assignments", server.partnerCreateRuleAssignment).Methods("POST")
	partnerAuthed.HandleFunc("/members/{memberId}/rule-assignments/{assignmentId}", server.partnerUpdateRuleAssignment).Methods("PATCH")
	partnerAuthed.HandleFunc("/ingest/transactions", server.partnerIngestTransaction).Methods("POST")
	partnerAuthed.HandleFunc("/transactions", server.partnerListTransactions).Methods("GET")
	partnerAuthed.HandleFunc("/transactions/{transactionEventId}/calculation", server.partnerGetTransactionCalculation).Methods("GET")
	partnerAuthed.HandleFunc("/jobs/process-transaction-events", server.partnerProcessTransactionEvents).Methods("POST")
	partnerAuthed.HandleFunc("/api-keys", server.createAPIKey).Methods("POST")
	partnerAuthed.HandleFunc("/api-keys", server.listAPIKeys).Methods("GET")
	partnerAuthed.HandleFunc("/api-keys/{keyId}", server.revokeAPIKey).Methods("DELETE")
	partnerAuthed.HandleFunc("/locations", server.createLocation).Methods("POST")
	partnerAuthed.HandleFunc("/locations", server.listLocations).Methods("GET")
	partnerAuthed.HandleFunc("/catalog-items", server.createCatalogItem).Methods("POST")
	partnerAuthed.HandleFunc("/catalog-items", server.listCatalogItems).Methods("GET")
	partnerAuthed.HandleFunc("/catalog-items/{catalogItemId}", server.updateCatalogItem).Methods("PATCH")
	partnerAuthed.HandleFunc("/redemptions", server.listRedemptions).Methods("GET")
	partnerAuthed.HandleFunc("/dashboard", server.partnerDashboard).Methods("GET")
	partnerAuthed.HandleFunc("/integration-connections", server.listIntegrationConnections).Methods("GET")
	partnerAuthed.HandleFunc("/integration-connections/square/oauth-start", server.startSquareOAuth).Methods("POST")
	partnerAuthed.HandleFunc("/integration-connections/square/oauth-callback", server.completeSquareOAuth).Methods("GET")
	partnerAuthed.HandleFunc("/integration-connections/{connectionId}/sync", server.syncIntegrationConnection).Methods("POST")
	partnerAuthed.HandleFunc("/campaigns", server.createCampaign).Methods("POST")
	partnerAuthed.HandleFunc("/campaigns", server.listCampaigns).Methods("GET")

	pos := r.PathPrefix("/pos/v1").Subrouter()
	pos.Use(server.requireAuth)
	pos.HandleFunc("/members/resolve", server.resolvePOSMember).Methods("POST")
	pos.HandleFunc("/manual-transactions", server.createManualTransaction).Methods("POST")
	pos.HandleFunc("/members/{memberId}/balance", server.getPOSMemberBalance).Methods("GET")
	pos.HandleFunc("/members/{memberId}/available-rewards", server.availableRewards).Methods("GET")
	pos.HandleFunc("/redemptions", server.createRedemption).Methods("POST")
	pos.HandleFunc("/redemptions/{redemptionId}/validate", server.validateRedemption).Methods("POST")
	pos.HandleFunc("/redemptions/{redemptionId}/capture", server.captureRedemption).Methods("POST")
	pos.HandleFunc("/redemptions/{redemptionId}/release", server.releaseRedemption).Methods("POST")
	return r
}

func (s Server) health(w http.ResponseWriter, _ *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]bool{"alive": true})
}

func (s Server) createPartner(w http.ResponseWriter, r *http.Request) {
	var body domain.PartnerRequest
	if !decode(w, r, &body) {
		return
	}
	partner, err := s.services.Partners.Create(r.Context(), body)
	respond(w, http.StatusCreated, partner, err)
}

func (s Server) listPartners(w http.ResponseWriter, r *http.Request) {
	partners, err := s.services.Partners.List(r.Context())
	respond(w, http.StatusOK, partners, err)
}

func (s Server) getPartner(w http.ResponseWriter, r *http.Request) {
	partner, err := s.services.Partners.GetByKey(r.Context(), vars(r)["partnerKey"])
	respond(w, http.StatusOK, partner, err)
}

func (s Server) createProgram(w http.ResponseWriter, r *http.Request) {
	var body domain.ProgramRequest
	if !decode(w, r, &body) {
		return
	}
	program, err := s.services.Programs.CreateProgram(r.Context(), vars(r)["partnerKey"], body)
	respond(w, http.StatusCreated, program, err)
}

func (s Server) listPrograms(w http.ResponseWriter, r *http.Request) {
	programs, err := s.services.Programs.ListPrograms(r.Context(), vars(r)["partnerKey"])
	respond(w, http.StatusOK, programs, err)
}

func (s Server) createRuleVersion(w http.ResponseWriter, r *http.Request) {
	var body domain.RuleVersionRequest
	if !decode(w, r, &body) {
		return
	}
	version, err := s.services.Programs.CreateRuleVersion(r.Context(), vars(r)["partnerKey"], vars(r)["programId"], body)
	respond(w, http.StatusCreated, version, err)
}

func (s Server) listRuleVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.services.Programs.ListRuleVersions(r.Context(), vars(r)["partnerKey"], vars(r)["programId"])
	respond(w, http.StatusOK, versions, err)
}

func (s Server) getRuleVersionReview(w http.ResponseWriter, r *http.Request) {
	review, err := s.services.Programs.GetRuleVersionReview(r.Context(), vars(r)["partnerKey"], vars(r)["programId"], vars(r)["versionId"])
	respond(w, http.StatusOK, review, err)
}

func (s Server) validateRuleVersion(w http.ResponseWriter, r *http.Request) {
	result := s.services.Programs.ValidateRuleVersion(r.Context(), vars(r)["partnerKey"], vars(r)["programId"], vars(r)["versionId"])
	respondWithJSON(w, http.StatusOK, result)
}

func (s Server) publishRuleVersion(w http.ResponseWriter, r *http.Request) {
	version, err := s.services.Programs.PublishRuleVersion(r.Context(), vars(r)["partnerKey"], vars(r)["programId"], vars(r)["versionId"])
	respond(w, http.StatusOK, version, err)
}

func (s Server) listRulePackages(w http.ResponseWriter, r *http.Request) {
	packages, err := s.services.Programs.ListRulePackages(r.Context(), vars(r)["partnerKey"], vars(r)["programId"])
	respond(w, http.StatusOK, packages, err)
}

func (s Server) createRulePackage(w http.ResponseWriter, r *http.Request) {
	var body domain.RuleVersionRequest
	if !decode(w, r, &body) {
		return
	}
	version, err := s.services.Programs.CreateRulePackage(r.Context(), vars(r)["partnerKey"], vars(r)["programId"], body)
	respond(w, http.StatusCreated, version, err)
}

func (s Server) createMember(w http.ResponseWriter, r *http.Request) {
	var body domain.MemberRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Members.CreateMember(r.Context(), vars(r)["partnerKey"], body)
	respond(w, http.StatusCreated, result, err)
}

func (s Server) listMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.services.Members.ListMembers(r.Context(), vars(r)["partnerKey"])
	respond(w, http.StatusOK, members, err)
}

func (s Server) getMember(w http.ResponseWriter, r *http.Request) {
	member, err := s.services.Members.GetMember(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"])
	respond(w, http.StatusOK, member, err)
}

func (s Server) getRewardsProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := s.services.Members.GetRewardsProfile(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"])
	respond(w, http.StatusOK, profile, err)
}

func (s Server) updateMemberEnrollment(w http.ResponseWriter, r *http.Request) {
	var body domain.EnrollmentRequest
	if !decode(w, r, &body) {
		return
	}
	err := s.services.Members.UpdateEnrollment(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"], body)
	respond(w, http.StatusOK, map[string]string{"status": "active", "programId": body.ProgramID}, err)
}

func (s Server) createRuleAssignment(w http.ResponseWriter, r *http.Request) {
	var body domain.MemberRuleAssignmentRequest
	if !decode(w, r, &body) {
		return
	}
	assignment, err := s.services.Members.CreateRuleAssignment(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"], body)
	respond(w, http.StatusCreated, assignment, err)
}

func (s Server) updateRuleAssignment(w http.ResponseWriter, r *http.Request) {
	var body domain.MemberRuleAssignmentUpdateRequest
	if !decode(w, r, &body) {
		return
	}
	assignment, err := s.services.Members.UpdateRuleAssignment(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"], vars(r)["assignmentId"], body)
	respond(w, http.StatusOK, assignment, err)
}

func (s Server) ingestTransaction(w http.ResponseWriter, r *http.Request) {
	var body domain.TransactionIngestRequest
	if !decode(w, r, &body) {
		return
	}
	event, err := s.services.Transactions.IngestTransaction(r.Context(), vars(r)["partnerKey"], body)
	respond(w, http.StatusOK, event, err)
}

func (s Server) getIngestedTransaction(w http.ResponseWriter, r *http.Request) {
	event, err := s.services.Transactions.GetTransaction(r.Context(), vars(r)["partnerKey"], vars(r)["transactionEventId"])
	respond(w, http.StatusOK, event, err)
}

func (s Server) listTransactions(w http.ResponseWriter, r *http.Request) {
	events, err := s.services.Transactions.ListTransactions(r.Context(), vars(r)["partnerKey"])
	respond(w, http.StatusOK, events, err)
}

func (s Server) getTransactionCalculation(w http.ResponseWriter, r *http.Request) {
	calculation, err := s.services.Transactions.GetCalculation(r.Context(), vars(r)["partnerKey"], vars(r)["transactionEventId"])
	respond(w, http.StatusOK, calculation, err)
}

func (s Server) processTransactionEvents(w http.ResponseWriter, r *http.Request) {
	result, err := s.services.RewardProcessing.ProcessTransactionEvents(r.Context())
	respond(w, http.StatusOK, result, err)
}

func (s Server) getMemberBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := s.services.Ledger.GetBalance(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"])
	respond(w, http.StatusOK, balance, err)
}

func (s Server) getMemberLedger(w http.ResponseWriter, r *http.Request) {
	entries, err := s.services.Ledger.GetLedger(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"])
	respond(w, http.StatusOK, entries, err)
}

func (s Server) createAdjustment(w http.ResponseWriter, r *http.Request) {
	var body domain.AdjustmentRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Ledger.CreateAdjustment(r.Context(), vars(r)["partnerKey"], vars(r)["memberId"], body)
	respond(w, http.StatusCreated, result, err)
}

func (s Server) generateLedgerLiabilityExport(w http.ResponseWriter, r *http.Request) {
	var body domain.ExportRequest
	if !decode(w, r, &body) {
		return
	}
	export, err := s.services.Reporting.GenerateLedgerLiabilityExport(r.Context(), body)
	respond(w, http.StatusOK, export, err)
}

func (s Server) listLedgerLiabilityExports(w http.ResponseWriter, r *http.Request) {
	exports, err := s.services.Reporting.ListLedgerLiabilityExports(r.Context(), vars(r)["partnerKey"])
	respond(w, http.StatusOK, exports, err)
}

func (s Server) partnerLogin(w http.ResponseWriter, r *http.Request) {
	var body domain.LoginRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Auth.Login(r.Context(), body)
	respond(w, http.StatusOK, result, err)
}

func (s Server) partnerLogout(w http.ResponseWriter, _ *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "signed_out"})
}

func (s Server) partnerMe(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, authFromContext(r.Context()))
}

func (s Server) partnerCreateProgram(w http.ResponseWriter, r *http.Request) {
	var body domain.ProgramRequest
	if !decode(w, r, &body) {
		return
	}
	program, err := s.services.Programs.CreateProgram(r.Context(), authFromContext(r.Context()).PartnerKey, body)
	respond(w, http.StatusCreated, program, err)
}

func (s Server) partnerListPrograms(w http.ResponseWriter, r *http.Request) {
	programs, err := s.services.Programs.ListPrograms(r.Context(), authFromContext(r.Context()).PartnerKey)
	respond(w, http.StatusOK, programs, err)
}

func (s Server) partnerCreateRuleVersion(w http.ResponseWriter, r *http.Request) {
	var body domain.RuleVersionRequest
	if !decode(w, r, &body) {
		return
	}
	version, err := s.services.Programs.CreateRuleVersion(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["programId"], body)
	respond(w, http.StatusCreated, version, err)
}

func (s Server) partnerListRuleVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.services.Programs.ListRuleVersions(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["programId"])
	respond(w, http.StatusOK, versions, err)
}

func (s Server) partnerGetRuleVersionReview(w http.ResponseWriter, r *http.Request) {
	review, err := s.services.Programs.GetRuleVersionReview(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["programId"], vars(r)["versionId"])
	respond(w, http.StatusOK, review, err)
}

func (s Server) partnerPublishRuleVersion(w http.ResponseWriter, r *http.Request) {
	version, err := s.services.Programs.PublishRuleVersion(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["programId"], vars(r)["versionId"])
	respond(w, http.StatusOK, version, err)
}

func (s Server) partnerListRulePackages(w http.ResponseWriter, r *http.Request) {
	packages, err := s.services.Programs.ListRulePackages(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["programId"])
	respond(w, http.StatusOK, packages, err)
}

func (s Server) partnerCreateRulePackage(w http.ResponseWriter, r *http.Request) {
	var body domain.RuleVersionRequest
	if !decode(w, r, &body) {
		return
	}
	version, err := s.services.Programs.CreateRulePackage(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["programId"], body)
	respond(w, http.StatusCreated, version, err)
}

func (s Server) partnerCreateMember(w http.ResponseWriter, r *http.Request) {
	var body domain.MemberRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Members.CreateMember(r.Context(), authFromContext(r.Context()).PartnerKey, body)
	respond(w, http.StatusCreated, result, err)
}

func (s Server) partnerListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.services.Members.ListMembers(r.Context(), authFromContext(r.Context()).PartnerKey)
	respond(w, http.StatusOK, members, err)
}

func (s Server) partnerGetRewardsProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := s.services.Members.GetRewardsProfile(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["memberId"])
	respond(w, http.StatusOK, profile, err)
}

func (s Server) partnerUpdateMemberEnrollment(w http.ResponseWriter, r *http.Request) {
	var body domain.EnrollmentRequest
	if !decode(w, r, &body) {
		return
	}
	err := s.services.Members.UpdateEnrollment(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["memberId"], body)
	respond(w, http.StatusOK, map[string]string{"status": "active", "programId": body.ProgramID}, err)
}

func (s Server) partnerCreateRuleAssignment(w http.ResponseWriter, r *http.Request) {
	var body domain.MemberRuleAssignmentRequest
	if !decode(w, r, &body) {
		return
	}
	assignment, err := s.services.Members.CreateRuleAssignment(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["memberId"], body)
	respond(w, http.StatusCreated, assignment, err)
}

func (s Server) partnerUpdateRuleAssignment(w http.ResponseWriter, r *http.Request) {
	var body domain.MemberRuleAssignmentUpdateRequest
	if !decode(w, r, &body) {
		return
	}
	assignment, err := s.services.Members.UpdateRuleAssignment(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["memberId"], vars(r)["assignmentId"], body)
	respond(w, http.StatusOK, assignment, err)
}

func (s Server) partnerIngestTransaction(w http.ResponseWriter, r *http.Request) {
	var body domain.TransactionIngestRequest
	if !decode(w, r, &body) {
		return
	}
	event, err := s.services.Transactions.IngestTransaction(r.Context(), authFromContext(r.Context()).PartnerKey, body)
	respond(w, http.StatusCreated, event, err)
}

func (s Server) partnerListTransactions(w http.ResponseWriter, r *http.Request) {
	events, err := s.services.Transactions.ListTransactions(r.Context(), authFromContext(r.Context()).PartnerKey)
	respond(w, http.StatusOK, events, err)
}

func (s Server) partnerGetTransactionCalculation(w http.ResponseWriter, r *http.Request) {
	calculation, err := s.services.Transactions.GetCalculation(r.Context(), authFromContext(r.Context()).PartnerKey, vars(r)["transactionEventId"])
	respond(w, http.StatusOK, calculation, err)
}

func (s Server) partnerProcessTransactionEvents(w http.ResponseWriter, r *http.Request) {
	result, err := s.services.RewardProcessing.ProcessTransactionEvents(r.Context())
	respond(w, http.StatusOK, result, err)
}

func (s Server) createAPIKey(w http.ResponseWriter, r *http.Request) {
	var body domain.APIKeyCreateRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Auth.CreateAPIKey(r.Context(), authFromContext(r.Context()), body)
	respond(w, http.StatusCreated, result, err)
}

func (s Server) listAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.services.Auth.ListAPIKeys(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, keys, err)
}

func (s Server) revokeAPIKey(w http.ResponseWriter, r *http.Request) {
	err := s.services.Auth.RevokeAPIKey(r.Context(), authFromContext(r.Context()), vars(r)["keyId"])
	respond(w, http.StatusOK, map[string]string{"status": "revoked"}, err)
}

func (s Server) createLocation(w http.ResponseWriter, r *http.Request) {
	var body domain.LocationRequest
	if !decode(w, r, &body) {
		return
	}
	location, err := s.services.Locations.CreateLocation(r.Context(), authFromContext(r.Context()), body)
	respond(w, http.StatusCreated, location, err)
}

func (s Server) listLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := s.services.Locations.ListLocations(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, locations, err)
}

func (s Server) createCatalogItem(w http.ResponseWriter, r *http.Request) {
	var body domain.CatalogItemRequest
	if !decode(w, r, &body) {
		return
	}
	item, err := s.services.Catalog.CreateCatalogItem(r.Context(), authFromContext(r.Context()), body)
	respond(w, http.StatusCreated, item, err)
}

func (s Server) listCatalogItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.services.Catalog.ListCatalogItems(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, items, err)
}

func (s Server) updateCatalogItem(w http.ResponseWriter, r *http.Request) {
	var body domain.CatalogItemRequest
	if !decode(w, r, &body) {
		return
	}
	item, err := s.services.Catalog.UpdateCatalogItem(r.Context(), authFromContext(r.Context()), vars(r)["catalogItemId"], body)
	respond(w, http.StatusOK, item, err)
}

func (s Server) resolvePOSMember(w http.ResponseWriter, r *http.Request) {
	var body domain.ResolveMemberRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Members.ResolveOrCreateMemberForPartnerID(r.Context(), authFromContext(r.Context()).PartnerID, body)
	respond(w, http.StatusOK, result, err)
}

func (s Server) createManualTransaction(w http.ResponseWriter, r *http.Request) {
	var body domain.ManualTransactionRequest
	if !decode(w, r, &body) {
		return
	}
	event, err := s.services.Transactions.IngestManualTransaction(r.Context(), authFromContext(r.Context()), body)
	respond(w, http.StatusCreated, event, err)
}

func (s Server) getPOSMemberBalance(w http.ResponseWriter, r *http.Request) {
	auth := authFromContext(r.Context())
	balance, err := s.services.Ledger.GetBalance(r.Context(), auth.PartnerKey, vars(r)["memberId"])
	respond(w, http.StatusOK, balance, err)
}

func (s Server) availableRewards(w http.ResponseWriter, r *http.Request) {
	items, err := s.services.Catalog.AvailableRewards(r.Context(), authFromContext(r.Context()), vars(r)["memberId"])
	respond(w, http.StatusOK, items, err)
}

func (s Server) createRedemption(w http.ResponseWriter, r *http.Request) {
	var body domain.RedemptionRequest
	if !decode(w, r, &body) {
		return
	}
	result, err := s.services.Redemptions.CreateRedemption(r.Context(), authFromContext(r.Context()), body)
	respond(w, http.StatusCreated, result, err)
}

func (s Server) validateRedemption(w http.ResponseWriter, r *http.Request) {
	result, err := s.services.Redemptions.ValidateRedemption(r.Context(), authFromContext(r.Context()), vars(r)["redemptionId"])
	respond(w, http.StatusOK, result, err)
}

func (s Server) captureRedemption(w http.ResponseWriter, r *http.Request) {
	result, err := s.services.Redemptions.CaptureRedemption(r.Context(), authFromContext(r.Context()), vars(r)["redemptionId"])
	respond(w, http.StatusOK, result, err)
}

func (s Server) releaseRedemption(w http.ResponseWriter, r *http.Request) {
	result, err := s.services.Redemptions.ReleaseRedemption(r.Context(), authFromContext(r.Context()), vars(r)["redemptionId"])
	respond(w, http.StatusOK, result, err)
}

func (s Server) listRedemptions(w http.ResponseWriter, r *http.Request) {
	redemptions, err := s.services.Redemptions.ListRedemptions(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, redemptions, err)
}

func (s Server) partnerDashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := s.services.Dashboard.Summary(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, summary, err)
}

func (s Server) listIntegrationConnections(w http.ResponseWriter, r *http.Request) {
	connections, err := s.services.Integrations.ListConnections(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, connections, err)
}

func (s Server) startSquareOAuth(w http.ResponseWriter, r *http.Request) {
	connection, err := s.services.Integrations.StartSquareOAuth(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusCreated, connection, err)
}

func (s Server) completeSquareOAuth(w http.ResponseWriter, r *http.Request) {
	connection, err := s.services.Integrations.CompleteSquareOAuth(r.Context(), authFromContext(r.Context()), r.URL.Query().Get("code"))
	respond(w, http.StatusCreated, connection, err)
}

func (s Server) syncIntegrationConnection(w http.ResponseWriter, r *http.Request) {
	connection, err := s.services.Integrations.SyncConnection(r.Context(), authFromContext(r.Context()), vars(r)["connectionId"])
	respond(w, http.StatusOK, connection, err)
}

func (s Server) createCampaign(w http.ResponseWriter, r *http.Request) {
	var body domain.CampaignRequest
	if !decode(w, r, &body) {
		return
	}
	campaign, err := s.services.Campaigns.CreateCampaign(r.Context(), authFromContext(r.Context()), body)
	respond(w, http.StatusCreated, campaign, err)
}

func (s Server) listCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := s.services.Campaigns.ListCampaigns(r.Context(), authFromContext(r.Context()))
	respond(w, http.StatusOK, campaigns, err)
}

type authContextKey struct{}

func (s Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.Header.Get("X-Paisa-API-Key")
		}
		token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
		auth, err := s.services.Auth.AuthenticateToken(r.Context(), token)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authContextKey{}, auth)))
	})
}

func authFromContext(ctx context.Context) domain.AuthContext {
	auth, _ := ctx.Value(authContextKey{}).(domain.AuthContext)
	return auth
}

func vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

func decode(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(target); err != nil && !errors.Is(err, io.EOF) {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func respond(w http.ResponseWriter, successCode int, payload interface{}, err error) {
	if err != nil {
		status := httpStatusForError(err)
		if status == http.StatusInternalServerError {
			log.Printf("request failed: %v", err)
		}
		respondWithError(w, status, publicErrorMessage(err))
		return
	}
	respondWithJSON(w, successCode, payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func httpStatusForError(err error) int {
	var appErr domain.AppError
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case domain.ErrorKindInvalid:
			return http.StatusBadRequest
		case domain.ErrorKindNotFound:
			return http.StatusNotFound
		case domain.ErrorKindConflict:
			return http.StatusConflict
		case domain.ErrorKindInvariant:
			return http.StatusUnprocessableEntity
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

func publicErrorMessage(err error) string {
	var appErr domain.AppError
	if errors.As(err, &appErr) {
		if appErr.Kind == domain.ErrorKindInternal {
			return "internal server error"
		}
		return appErr.Error()
	}
	return "internal server error"
}
