package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"accts-api/domain"
	"accts-api/ports"

	"github.com/gorilla/mux"
)

type Services struct {
	Partners         ports.PartnerService
	Programs         ports.ProgramService
	Members          ports.MemberService
	Transactions     ports.TransactionIngestionService
	RewardProcessing ports.RewardProcessingService
	Ledger           ports.LedgerService
	Reporting        ports.ReportingService
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
