package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (newApiRequest *ApiAccess) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Check if username and password are provided and present in the database
	var body LoginRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		log.Printf("Error unmarshalling request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if body.Username == "" || body.Password == "" {
		log.Printf("Username or password not provided")
		respondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}
	//Check if the user exists in the database
	var userAccountId string
	var hashedPassword string
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = newApiRequest.AccountsDB.QueryRowContext(ctx, "SELECT accountid, password FROM paisa.users WHERE username=$1", body.Username).Scan(&userAccountId, &hashedPassword)
	if err != nil {
		log.Printf("Error checking user existence: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	if userAccountId == "" {
		log.Printf("User does not exist")
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	// Compare password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(body.Password))
	if err != nil {
		log.Printf("Password does not match for user %s", body.Username)
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	//Generate a session id for the user
	session, err := newApiRequest.generateSessionForAccount(userAccountId, r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create session")
		return
	}
	// return the session id to the user
	respondWithJSON(w, http.StatusOK, session)
}

// Removed RegisterUser from this file. Now implemented in api.go.

// Authorization middleware to check session
func (newApiRequest *ApiAccess) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("Authorization")
		if sessionId == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing session token")
			return
		}
		var accountId string
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		err := newApiRequest.AccountsDB.QueryRowContext(ctx, "SELECT accountId FROM paisa.sessions WHERE sessionId=$1", sessionId).Scan(&accountId)
		if err != nil || accountId == "" {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired session token")
			return
		}
		// Optionally, set accountId in context for downstream handlers
		next.ServeHTTP(w, r)
	})
}

func (newApiRequest *ApiAccess) generateSessionForAccount(accountId string, ctx context.Context) (Session, error) {
	generateSessionId := uuid.NewString()
	session := Session{
		SessionId: generateSessionId,
		AccountId: accountId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := newApiRequest.AccountsDB.ExecContext(ctx, "INSERT INTO paisa.sessions (sessionId, accountId) VALUES ($1, $2)", session.SessionId, accountId)
	if err != nil {
		log.Printf("Error inserting into postgres: %s", err)
		return Session{}, err
	}
	return session, nil
}
