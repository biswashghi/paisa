package api

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

/*
---------------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------------
*/

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
type ContextKey string

// ContextKeyRequestID is the ContextKey for RequestID
const ContextKeyRequestID ContextKey = "requestID"

// Accounts Postgres Collection Model
type AccountSnapshot struct {
	AccountId      string          `json:"accountId"`
	RewardsBalance decimal.Decimal `json:"rewardsBalance"`
}

// Transaction History Postgres Collection Model
type TransactionHistory struct {
	TransactionId string          `json:"transactionId"`
	AccountId     string          `json:"accountId"`
	Description   string          `json:"description"`
	MerchantCode  string          `json:"merchantCode"`
	Increment     decimal.Decimal `json:"increment"`
	CreatedAt     string          `json:"createdAt"`
}

// LoginRequest is the request body for the login endpoint
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionId string `json:"sessionId"`
	AccountId string `json:"accountId"`
}

// RegisterRequest is the request body for the register endpoint
// and RegisterResponse for the response
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	AccountId string `json:"accountId"`
	Message   string `json:"message"`
}

// Session is the session object
type Session struct {
	SessionId string    `json:"sessionId"`
	AccountId string    `json:"accountId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
---------------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------------
---------------------------------------------------------------------------------------------------------------------------------------
*/
// Create a custom Env struct which holds a connection pool.
type ApiAccess struct {
	AccountsDB *sql.DB
}

func NewApiAccess(ctx context.Context) *ApiAccess {
	api := ApiAccess{}
	api.CreateClient(ctx)
	return &api
}

const (
	host     = "localhost"
	port     = 5243
	user     = "project"
	password = "project"
	dbname   = "project"
)

func (api *ApiAccess) CreateClient(ctx context.Context) {
	/*
	   Connect to my cluster
	*/
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	api.AccountsDB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = api.AccountsDB.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

/*
---------------------------------------------------------------------------------------------------------------------------------------
*/

type CreateAccountRequest struct {
	AccountId string `json:"accountId"`
}

type CreateAccountResponse struct {
	AccountId      string `json:"accountId"`
	RewardsBalance string `json:"rewardsBalance"`
}

/*
---------------------------------------------------------------------------------------------------------------------------------------
*/

type UpdateBalanceRequest struct {
	Increment    decimal.Decimal `json:"balance"`
	Description  string          `json:"description"`
	MerchantCode string          `json:"merchantCode"`
}
