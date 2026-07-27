package application

import (
	"context"
	"strings"
	"time"

	"accts-api/domain"
	"accts-api/ports"
)

type AuthService struct {
	app app
}

func (s AuthService) OnboardPartner(ctx context.Context, body domain.PartnerOnboardRequest) (domain.PartnerOnboardResult, error) {
	partnerKey := strings.TrimSpace(body.PartnerKey)
	if partnerKey == "" {
		return domain.PartnerOnboardResult{}, domain.InvalidError("partnerKey is required")
	}
	partnerName := strings.TrimSpace(body.PartnerName)
	if partnerName == "" {
		partnerName = titleizePartnerKey(partnerKey)
	}
	email := strings.TrimSpace(body.AdminEmail)
	if email == "" {
		return domain.PartnerOnboardResult{}, domain.InvalidError("adminEmail is required")
	}
	name := strings.TrimSpace(body.AdminName)
	if name == "" {
		name = "Partner Admin"
	}
	passwordHash, err := domain.HashPassword(body.AdminPassword)
	if err != nil {
		return domain.PartnerOnboardResult{}, err
	}

	var result domain.PartnerOnboardResult
	err = s.app.uow.WithinTx(ctx, func(txCtx context.Context, stores ports.StoreSet) error {
		partner, err := stores.Partners.GetByKey(txCtx, partnerKey)
		if err != nil {
			if !domain.IsErrorKind(err, domain.ErrorKindNotFound) {
				return err
			}
			partner, err = stores.Partners.Create(txCtx, domain.PartnerRequest{
				PartnerKey: partnerKey,
				Name:       partnerName,
			})
			if err != nil {
				return err
			}
		}
		if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
			return err
		}

		user, err := stores.Auth.UpsertPartnerUserWithPassword(txCtx, partner.ID, email, name, passwordHash)
		if err != nil {
			return err
		}
		result = domain.PartnerOnboardResult{Partner: partner, User: user}
		return nil
	})
	if err != nil {
		return domain.PartnerOnboardResult{}, err
	}
	result.User.PasswordHash = ""
	return result, nil
}

func (s AuthService) Login(ctx context.Context, body domain.LoginRequest) (domain.LoginResult, error) {
	email := strings.TrimSpace(body.Email)
	if email == "" {
		return domain.LoginResult{}, domain.InvalidError("email is required")
	}
	if strings.TrimSpace(body.Password) == "" {
		return domain.LoginResult{}, domain.InvalidError("password is required")
	}

	partner, user, err := s.app.stores.Auth.PartnerUserByEmail(ctx, email)
	if err != nil {
		if domain.IsErrorKind(err, domain.ErrorKindNotFound) {
			return domain.LoginResult{}, domain.InvalidError("invalid email or password")
		}
		return domain.LoginResult{}, err
	}
	if !domain.VerifyPassword(body.Password, user.PasswordHash) {
		return domain.LoginResult{}, domain.InvalidError("invalid email or password")
	}
	if err := domain.EnsureActiveStatus("partner", partner.Status); err != nil {
		return domain.LoginResult{}, err
	}
	if err := domain.EnsureActiveStatus("partner user", user.Status); err != nil {
		return domain.LoginResult{}, err
	}

	token := domain.GenerateToken("ps_session")
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := s.app.stores.Auth.CreateSession(ctx, partner.ID, user.ID, domain.HashSecret(token), expiresAt); err != nil {
		return domain.LoginResult{}, err
	}
	auth := domain.AuthContext{
		PartnerID:     partner.ID,
		PartnerKey:    partner.PartnerKey,
		PartnerName:   partner.Name,
		ActorType:     "partner_user",
		ActorID:       user.ID,
		Scopes:        []string{"partner:read", "partner:write", "pos:write"},
		Authenticated: time.Now().UTC(),
	}
	user.PasswordHash = ""
	return domain.LoginResult{Token: token, Auth: auth, Partner: partner, User: user}, nil
}

func (s AuthService) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return nil
	}
	return s.app.stores.Auth.RevokeSessionHash(ctx, domain.HashSecret(token))
}

func (s AuthService) AuthenticateToken(ctx context.Context, token string) (domain.AuthContext, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return domain.AuthContext{}, domain.InvalidError("authentication token is required")
	}
	hash := domain.HashSecret(token)
	auth, err := s.app.stores.Auth.AuthBySessionHash(ctx, hash)
	if err == nil {
		return auth, nil
	}
	if !domain.IsErrorKind(err, domain.ErrorKindNotFound) {
		return domain.AuthContext{}, err
	}
	auth, err = s.app.stores.Auth.AuthByAPIKeyHash(ctx, hash)
	if err != nil {
		return domain.AuthContext{}, err
	}
	_ = s.app.stores.Auth.TouchAPIKey(ctx, auth.ActorID)
	return auth, nil
}

func (s AuthService) CreateAPIKey(ctx context.Context, auth domain.AuthContext, body domain.APIKeyCreateRequest) (domain.APIKeyCreateResult, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "Cashier API key"
	}
	scopes := body.Scopes
	if len(scopes) == 0 {
		scopes = []string{"partner:read", "pos:write", "ingest:write"}
	}
	token := domain.GenerateToken("pk_paisa")
	prefix := token
	if len(prefix) > 18 {
		prefix = prefix[:18]
	}
	key, err := s.app.stores.Auth.CreateAPIKey(ctx, auth.PartnerID, name, prefix, domain.HashSecret(token), scopes)
	if err != nil {
		return domain.APIKeyCreateResult{}, err
	}
	return domain.APIKeyCreateResult{APIKey: key, Token: token}, nil
}

func (s AuthService) ListAPIKeys(ctx context.Context, auth domain.AuthContext) ([]domain.APIKey, error) {
	return s.app.stores.Auth.ListAPIKeys(ctx, auth.PartnerID)
}

func (s AuthService) RevokeAPIKey(ctx context.Context, auth domain.AuthContext, keyID string) error {
	return s.app.stores.Auth.RevokeAPIKey(ctx, auth.PartnerID, keyID)
}

func titleizePartnerKey(key string) string {
	parts := strings.Split(key, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
