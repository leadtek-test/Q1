package adapters

import (
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/auth"
)

type JWTManager struct {
	secret []byte
	expire time.Duration

	signToken func(claims customClaims) (string, error)
}

type customClaims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var _ auth.TokenManager = (*JWTManager)(nil)

func NewTokenManagerRepositoryJWT(secret string, expire time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		expire: expire,
	}
}

func (m *JWTManager) Generate(userID uint, username string) (string, time.Time, error) {
	if len(m.secret) == 0 {
		return "", time.Time{}, commonerrors.New(consts.ErrnoAuthJWTSecretNotConfigured)
	}
	if m.expire <= 0 {
		return "", time.Time{}, commonerrors.New(consts.ErrnoAuthJWTInvalidExpireConfig)
	}

	now := time.Now().UTC()
	expireAt := now.Add(m.expire)

	claims := customClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	token, err := m.sign(claims)
	if err != nil {
		return "", time.Time{}, commonerrors.NewWithError(consts.ErrnoAuthJWTSignFailed, err)
	}

	return token, expireAt, nil
}

func (m *JWTManager) Parse(token string) (auth.Claims, error) {
	if len(m.secret) == 0 {
		return auth.Claims{}, commonerrors.New(consts.ErrnoAuthJWTSecretNotConfigured)
	}

	parsed, err := jwt.ParseWithClaims(token, &customClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			alg := "<nil>"
			if t.Method != nil {
				alg = t.Method.Alg()
			}
			return nil, fmt.Errorf("unexpected signing method: %s", alg)
		}
		return m.secret, nil
	})
	if err != nil {
		return auth.Claims{}, parseError(err)
	}

	claims, ok := parsed.Claims.(*customClaims)
	if !ok || claims == nil || claims.ExpiresAt == nil {
		return auth.Claims{}, commonerrors.New(consts.ErrnoAuthJWTInvalidClaims)
	}

	return auth.Claims{
		UserID:    claims.UserID,
		Username:  claims.Username,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func parseError(err error) error {
	switch {
	case stderrors.Is(err, jwt.ErrTokenMalformed):
		return commonerrors.NewWithError(consts.ErrnoAuthJWTMalformedToken, err)
	case stderrors.Is(err, jwt.ErrTokenSignatureInvalid):
		return commonerrors.NewWithError(consts.ErrnoAuthJWTSignatureInvalid, err)
	case stderrors.Is(err, jwt.ErrTokenExpired):
		return commonerrors.NewWithError(consts.ErrnoAuthJWTExpiredToken, err)
	default:
		return commonerrors.NewWithError(consts.ErrnoAuthJWTParseFailed, err)
	}
}

func (m *JWTManager) sign(claims customClaims) (string, error) {
	if m.signToken != nil {
		return m.signToken(claims)
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}
