package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"accts-api/domain"

	"github.com/lib/pq"
)

func (s AuthStore) UpsertPartnerUserWithPassword(ctx context.Context, partnerID, email, name, passwordHash string) (domain.PartnerUser, error) {
	var user domain.PartnerUser
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.partner_users (partner_id, email, name, password_hash)
		VALUES ($1, lower($2), $3, $4)
		ON CONFLICT (partner_id, email)
		DO UPDATE SET name = EXCLUDED.name, password_hash = EXCLUDED.password_hash, status = 'active', updated_at = now()
		RETURNING id, partner_id, email, name, password_hash, role, status, created_at, updated_at`,
		partnerID, email, name, passwordHash,
	).Scan(&user.ID, &user.PartnerID, &user.Email, &user.Name, &user.PasswordHash, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	return user, AppErrorFromDB(err)
}

func (s AuthStore) PartnerUserByEmail(ctx context.Context, email string) (domain.Partner, domain.PartnerUser, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT p.id, p.partner_key, p.name, p.status, p.created_at, p.updated_at,
		       pu.id, pu.partner_id, pu.email, pu.name, COALESCE(pu.password_hash, ''), pu.role, pu.status, pu.created_at, pu.updated_at
		FROM paisa.partner_users pu
		JOIN paisa.partners p ON p.id = pu.partner_id
		WHERE lower(pu.email) = lower($1)
		  AND pu.status = 'active'
		  AND p.status = 'active'
		ORDER BY pu.created_at ASC
		LIMIT 2`, strings.TrimSpace(email))
	if err != nil {
		return domain.Partner{}, domain.PartnerUser{}, AppErrorFromDB(err)
	}
	defer rows.Close()

	type identity struct {
		partner domain.Partner
		user    domain.PartnerUser
	}
	identities := []identity{}
	for rows.Next() {
		var item identity
		if err := rows.Scan(
			&item.partner.ID, &item.partner.PartnerKey, &item.partner.Name, &item.partner.Status, &item.partner.CreatedAt, &item.partner.UpdatedAt,
			&item.user.ID, &item.user.PartnerID, &item.user.Email, &item.user.Name, &item.user.PasswordHash, &item.user.Role, &item.user.Status, &item.user.CreatedAt, &item.user.UpdatedAt,
		); err != nil {
			return domain.Partner{}, domain.PartnerUser{}, AppErrorFromDB(err)
		}
		identities = append(identities, item)
	}
	if err := rows.Err(); err != nil {
		return domain.Partner{}, domain.PartnerUser{}, AppErrorFromDB(err)
	}
	if len(identities) == 0 {
		return domain.Partner{}, domain.PartnerUser{}, AppErrorFromDB(sql.ErrNoRows)
	}
	if len(identities) > 1 {
		return domain.Partner{}, domain.PartnerUser{}, domain.InvalidError("email belongs to multiple partner workspaces")
	}
	return identities[0].partner, identities[0].user, nil
}

func (s AuthStore) CreateSession(ctx context.Context, partnerID, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO paisa.sessions (partner_id, partner_user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`, partnerID, userID, tokenHash, expiresAt)
	return AppErrorFromDB(err)
}

func (s AuthStore) RevokeSessionHash(ctx context.Context, tokenHash string) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.sessions
		SET status = 'revoked', updated_at = now()
		WHERE token_hash = $1`, tokenHash)
	return AppErrorFromDB(err)
}

func (s AuthStore) AuthBySessionHash(ctx context.Context, tokenHash string) (domain.AuthContext, error) {
	var auth domain.AuthContext
	err := s.q.QueryRowContext(ctx, `
		SELECT p.id, p.partner_key, p.name, 'partner_user', pu.id, ARRAY['partner:read','partner:write','pos:write']::text[]
		FROM paisa.sessions sess
		JOIN paisa.partners p ON p.id = sess.partner_id
		JOIN paisa.partner_users pu ON pu.id = sess.partner_user_id
		WHERE sess.token_hash = $1 AND sess.status = 'active' AND sess.expires_at > now()
			AND p.status = 'active' AND pu.status = 'active'
			AND COALESCE(pu.password_hash, '') <> ''`,
		tokenHash,
	).Scan(&auth.PartnerID, &auth.PartnerKey, &auth.PartnerName, &auth.ActorType, &auth.ActorID, pq.Array(&auth.Scopes))
	auth.Authenticated = time.Now().UTC()
	return auth, AppErrorFromDB(err)
}

func (s AuthStore) CreateAPIKey(ctx context.Context, partnerID, name, keyPrefix, secretHash string, scopes []string) (domain.APIKey, error) {
	var key domain.APIKey
	var lastUsed, revoked sql.NullTime
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.api_keys (partner_id, name, key_prefix, secret_hash, scopes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, partner_id, name, key_prefix, scopes, status, last_used_at, created_at, revoked_at`,
		partnerID, name, keyPrefix, secretHash, pq.Array(scopes),
	).Scan(&key.ID, &key.PartnerID, &key.Name, &key.KeyPrefix, pq.Array(&key.Scopes), &key.Status, &lastUsed, &key.CreatedAt, &revoked)
	if lastUsed.Valid {
		key.LastUsedAt = &lastUsed.Time
	}
	if revoked.Valid {
		key.RevokedAt = &revoked.Time
	}
	return key, AppErrorFromDB(err)
}

func (s AuthStore) AuthByAPIKeyHash(ctx context.Context, secretHash string) (domain.AuthContext, error) {
	var auth domain.AuthContext
	err := s.q.QueryRowContext(ctx, `
		SELECT p.id, p.partner_key, p.name, 'api_key', ak.id, ak.scopes
		FROM paisa.api_keys ak
		JOIN paisa.partners p ON p.id = ak.partner_id
		WHERE ak.secret_hash = $1 AND ak.status = 'active' AND p.status = 'active'`,
		secretHash,
	).Scan(&auth.PartnerID, &auth.PartnerKey, &auth.PartnerName, &auth.ActorType, &auth.ActorID, pq.Array(&auth.Scopes))
	auth.Authenticated = time.Now().UTC()
	return auth, AppErrorFromDB(err)
}

func (s AuthStore) TouchAPIKey(ctx context.Context, apiKeyID string) error {
	_, err := s.q.ExecContext(ctx, `UPDATE paisa.api_keys SET last_used_at = now() WHERE id = $1`, apiKeyID)
	return AppErrorFromDB(err)
}

func (s AuthStore) ListAPIKeys(ctx context.Context, partnerID string) ([]domain.APIKey, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, name, key_prefix, scopes, status, last_used_at, created_at, revoked_at
		FROM paisa.api_keys
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	keys := []domain.APIKey{}
	for rows.Next() {
		var key domain.APIKey
		var lastUsed, revoked sql.NullTime
		if err := rows.Scan(&key.ID, &key.PartnerID, &key.Name, &key.KeyPrefix, pq.Array(&key.Scopes), &key.Status, &lastUsed, &key.CreatedAt, &revoked); err != nil {
			return nil, AppErrorFromDB(err)
		}
		if lastUsed.Valid {
			key.LastUsedAt = &lastUsed.Time
		}
		if revoked.Valid {
			key.RevokedAt = &revoked.Time
		}
		keys = append(keys, key)
	}
	return keys, AppErrorFromDB(rows.Err())
}

func (s AuthStore) RevokeAPIKey(ctx context.Context, partnerID, keyID string) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.api_keys
		SET status = 'revoked', revoked_at = now()
		WHERE partner_id = $1 AND id = $2`, partnerID, keyID)
	return AppErrorFromDB(err)
}
