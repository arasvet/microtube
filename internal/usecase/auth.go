package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/arasvet/microtube/internal/config"
	"golang.org/x/crypto/bcrypt"

	"github.com/arasvet/microtube/internal/repo"
	"github.com/google/uuid"
)

const (
	defaultJwtTTL = 30 * time.Minute
)

var (
	ErrInvalidCredentials            = errors.New("invalid credentials")
	ErrInvalidEmailOrPass            = errors.New("invalid email/pass")
	ErrInvalidToken                  = errors.New("invalid token")
	ErrInvalidSignInMethodNotAllowed = errors.New("jwt: invalid signing method")
	ErrUnexpectedMethod              = errors.New("jwt: unexpected signing method")
	ErrSecretNotSet                  = errors.New("secret not set")
)

type AuthUC struct {
	cfg    config.Config
	store  repo.Store
	secret string
}

func NewAuthUC(cfg config.Config, store repo.Store, secret string) *AuthUC {
	return &AuthUC{
		cfg:    cfg,
		store:  store,
		secret: secret,
	}
}

func (uc *AuthUC) Register(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" {
		return "", ErrInvalidEmailOrPass
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	id := uuid.NewString()

	return uc.store.CreateUser(ctx, id, email, string(hash))
}

func (uc *AuthUC) Login(ctx context.Context, email, password string) (string, error) {
	id, hash, err := uc.store.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	return uc.issueJWT(id, uc.cfg.AuthTTL)
}

func (uc *AuthUC) issueJWT(sub string, ttl time.Duration) (string, error) {
	if uc.secret == "" {
		return "", ErrSecretNotSet
	}

	if ttl <= 0 {
		ttl = defaultJwtTTL
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   sub,                                            // sub
		IssuedAt:  jwt.NewNumericDate(now),                        // iat
		NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)), // nbf (чуть-чуть leeway назад)
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),               // exp
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Явно фиксируем метод подписи, чтобы не поймать alg none / подмену.
	if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return "", ErrUnexpectedMethod
	}

	signed, err := token.SignedString(uc.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

// ParseAndValidateJWT — парсинг + валидация времени с leeway.
func (uc *AuthUC) ParseAndValidateJWT(tokenStr string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	keyfunc := func(t *jwt.Token) (interface{}, error) {
		// Защищаемся от подмены алгоритма.
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidSignInMethodNotAllowed
		}

		return uc.secret, nil
	}

	parsed, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		keyfunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(60*time.Second),
	)
	if err != nil {
		return nil, err
	}

	if !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
