package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"accts-api/domain"
	"accts-api/ports"

	"github.com/lib/pq"
)

type Queryer interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

type PartnerStore struct {
	q Queryer
}

type AuthStore struct {
	q Queryer
}

type ProgramStore struct {
	q Queryer
}

type MemberStore struct {
	q Queryer
}

type TransactionStore struct {
	q Queryer
}

type RuleStore struct {
	q Queryer
}

type RewardCalculationStore struct {
	q Queryer
}

type LedgerStore struct {
	q Queryer
}

type ReportingStore struct {
	q Queryer
}

type LocationStore struct {
	q Queryer
}

type CatalogStore struct {
	q Queryer
}

type RedemptionStore struct {
	q Queryer
}

type IntegrationStore struct {
	q Queryer
}

type CampaignStore struct {
	q Queryer
}

func NewStoreSet(q Queryer) ports.StoreSet {
	return ports.StoreSet{
		Auth:               AuthStore{q: q},
		Partners:           PartnerStore{q: q},
		Programs:           ProgramStore{q: q},
		Members:            MemberStore{q: q},
		Transactions:       TransactionStore{q: q},
		Rules:              RuleStore{q: q},
		RewardCalculations: RewardCalculationStore{q: q},
		Ledger:             LedgerStore{q: q},
		Reporting:          ReportingStore{q: q},
		Locations:          LocationStore{q: q},
		Catalog:            CatalogStore{q: q},
		Redemptions:        RedemptionStore{q: q},
		Integrations:       IntegrationStore{q: q},
		Campaigns:          CampaignStore{q: q},
	}
}

func normalizeDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func AppErrorFromDB(err error) error {
	if err == nil {
		return nil
	}
	var appErr domain.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AppError{Kind: domain.ErrorKindNotFound, Message: "not found", Err: err}
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "duplicate key"):
		return domain.AppError{Kind: domain.ErrorKindConflict, Message: "duplicate record", Err: err}
	case strings.Contains(msg, "violates foreign key"):
		return domain.AppError{Kind: domain.ErrorKindInvalid, Message: "referenced record does not exist", Err: err}
	default:
		return domain.AppError{Kind: domain.ErrorKindInternal, Message: "database error", Err: err}
	}
}

func mustJSON(value interface{}) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return data
}

func scanJSON(data []byte) domain.JSONMap {
	if len(data) == 0 {
		return domain.JSONMap{}
	}
	var out domain.JSONMap
	if err := json.Unmarshal(data, &out); err != nil {
		return domain.JSONMap{"raw": string(data)}
	}
	return out
}

func pqArray(values []string) interface{} {
	return pq.Array(values)
}
