package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stakestock/backend/internal/models"
)

const sessionCookieName = "session"
const sessionDuration = 7 * 24 * time.Hour

type SessionStore struct {
	pool          *pgxpool.Pool
	secure        bool
	startingCoins string
}

func NewSessionStore(pool *pgxpool.Pool, secure bool, startingCoins string) *SessionStore {
	return &SessionStore{pool: pool, secure: secure, startingCoins: startingCoins}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// UpsertUser creates or updates a user from Google OAuth info and returns the user.
func (s *SessionStore) UpsertUser(ctx context.Context, googleSub, email string) (*models.User, error) {
	userID := uuid.New()
	var user models.User

	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (id, google_sub, email, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (google_sub) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, google_sub, email, username, created_at
	`, userID, googleSub, email).Scan(&user.ID, &user.GoogleSub, &user.Email, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	// Ensure balance row exists
	coins, err := decimal.NewFromString(s.startingCoins)
	if err != nil {
		coins = decimal.NewFromInt(1000)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO balances (user_id, cash, cash_locked, special_coins) VALUES ($1, $2, 0, 10)
		ON CONFLICT (user_id) DO NOTHING
	`, user.ID, coins)
	if err != nil {
		return nil, fmt.Errorf("ensure balance: %w", err)
	}

	return &user, nil
}

// CreateSession inserts a new session and sets the cookie.
func (s *SessionStore) CreateSession(ctx context.Context, w http.ResponseWriter, userID uuid.UUID) error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(sessionDuration)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	sameSite := http.SameSiteLaxMode
	if s.secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: sameSite,
	})
	return nil
}

// DeleteSession removes a session from the DB and clears the cookie.
func (s *SessionStore) DeleteSession(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil {
		_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE id=$1`, c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookieName,
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}

type contextKey string

const ctxUserKey contextKey = "user"

// GetUserFromRequest resolves the session cookie to a user.
func (s *SessionStore) GetUserFromRequest(r *http.Request) (*models.User, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}
	return s.GetUserByToken(r.Context(), c.Value)
}

func (s *SessionStore) GetUserByToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.google_sub, u.email, u.username, u.created_at
		FROM sessions sess
		JOIN users u ON u.id = sess.user_id
		WHERE sess.id = $1 AND sess.expires_at > now()
	`, token).Scan(&user.ID, &user.GoogleSub, &user.Email, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}
	return &user, nil
}

func (s *SessionStore) IsSecure() bool { return s.secure }

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(ctxUserKey).(*models.User)
	return u
}
