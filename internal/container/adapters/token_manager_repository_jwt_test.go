package adapters

import (
	stderrors "errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
)

func TestNewJWTManager(t *testing.T) {
	m := NewTokenManagerRepositoryJWT("secret", time.Minute)
	if string(m.secret) != "secret" {
		t.Fatalf("unexpected secret: %q", string(m.secret))
	}
	if m.expire != time.Minute {
		t.Fatalf("unexpected expire: %v", m.expire)
	}
}

func TestJWTManagerGenerate(t *testing.T) {
	t.Run("secret not configured", func(t *testing.T) {
		m := NewTokenManagerRepositoryJWT("", time.Minute)
		token, expireAt, err := m.Generate(1, "alice")
		if token != "" {
			t.Fatalf("token should be empty")
		}
		if !expireAt.IsZero() {
			t.Fatalf("expireAt should be zero")
		}
		assertErrno(t, err, consts.ErrnoAuthJWTSecretNotConfigured)
	})

	t.Run("invalid expire config", func(t *testing.T) {
		m := NewTokenManagerRepositoryJWT("secret", 0)
		token, expireAt, err := m.Generate(1, "alice")
		if token != "" {
			t.Fatalf("token should be empty")
		}
		if !expireAt.IsZero() {
			t.Fatalf("expireAt should be zero")
		}
		assertErrno(t, err, consts.ErrnoAuthJWTInvalidExpireConfig)
	})

	t.Run("success", func(t *testing.T) {
		m := NewTokenManagerRepositoryJWT("secret", 2*time.Minute)
		start := time.Now().UTC()
		token, expireAt, err := m.Generate(7, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatalf("token should not be empty")
		}
		if expireAt.Before(start.Add(time.Minute)) || expireAt.After(start.Add(3*time.Minute)) {
			t.Fatalf("expireAt out of expected range: %v", expireAt)
		}
	})

	t.Run("sign failed", func(t *testing.T) {
		m := &JWTManager{
			secret: []byte("secret"),
			expire: time.Minute,
			signToken: func(claims customClaims) (string, error) {
				return "", stderrors.New("sign error")
			},
		}

		token, expireAt, err := m.Generate(1, "alice")
		if token != "" {
			t.Fatalf("token should be empty")
		}
		if !expireAt.IsZero() {
			t.Fatalf("expireAt should be zero")
		}
		assertErrno(t, err, consts.ErrnoAuthJWTSignFailed)
	})
}

func TestJWTManagerParse(t *testing.T) {
	t.Run("secret not configured", func(t *testing.T) {
		m := NewTokenManagerRepositoryJWT("", time.Minute)
		_, err := m.Parse("x")
		assertErrno(t, err, consts.ErrnoAuthJWTSecretNotConfigured)
	})

	t.Run("malformed token", func(t *testing.T) {
		m := NewTokenManagerRepositoryJWT("secret", time.Minute)
		_, err := m.Parse("not-a-jwt")
		assertErrno(t, err, consts.ErrnoAuthJWTMalformedToken)
	})

	t.Run("signature invalid", func(t *testing.T) {
		issuer := NewTokenManagerRepositoryJWT("secret-1", time.Minute)
		token, _, err := issuer.Generate(3, "bob")
		if err != nil {
			t.Fatalf("generate token failed: %v", err)
		}

		verifier := NewTokenManagerRepositoryJWT("secret-2", time.Minute)
		_, err = verifier.Parse(token)
		assertErrno(t, err, consts.ErrnoAuthJWTSignatureInvalid)
	})

	t.Run("expired token", func(t *testing.T) {
		token := signToken(t, []byte("secret"), customClaims{
			UserID:   9,
			Username: "expired-user",
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-time.Hour)),
			},
		})

		m := NewTokenManagerRepositoryJWT("secret", time.Minute)
		_, err := m.Parse(token)
		assertErrno(t, err, consts.ErrnoAuthJWTExpiredToken)
	})

	t.Run("invalid claims without expiresAt", func(t *testing.T) {
		token := signToken(t, []byte("secret"), customClaims{
			UserID:   11,
			Username: "no-exp",
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			},
		})

		m := NewTokenManagerRepositoryJWT("secret", time.Minute)
		_, err := m.Parse(token)
		assertErrno(t, err, consts.ErrnoAuthJWTInvalidClaims)
	})

	t.Run("parse failed default branch", func(t *testing.T) {
		token := signToken(t, []byte("secret"), customClaims{
			UserID:   12,
			Username: "not-yet-valid",
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				NotBefore: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
			},
		})

		m := NewTokenManagerRepositoryJWT("secret", time.Minute)
		_, err := m.Parse(token)
		assertErrno(t, err, consts.ErrnoAuthJWTParseFailed)
	})

	t.Run("unexpected signing method", func(t *testing.T) {
		token := signTokenWithMethod(t, []byte("secret"), customClaims{
			UserID:   13,
			Username: "wrong-alg",
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			},
		}, jwt.SigningMethodHS512)

		m := NewTokenManagerRepositoryJWT("secret", time.Minute)
		_, err := m.Parse(token)
		assertErrno(t, err, consts.ErrnoAuthJWTParseFailed)
	})

	t.Run("success", func(t *testing.T) {
		m := NewTokenManagerRepositoryJWT("secret", 5*time.Minute)
		token, expireAt, err := m.Generate(42, "charlie")
		if err != nil {
			t.Fatalf("generate token failed: %v", err)
		}

		claims, err := m.Parse(token)
		if err != nil {
			t.Fatalf("parse token failed: %v", err)
		}
		if claims.UserID != 42 {
			t.Fatalf("unexpected userID: %d", claims.UserID)
		}
		if claims.Username != "charlie" {
			t.Fatalf("unexpected username: %s", claims.Username)
		}
		if delta := claims.ExpiresAt.Sub(expireAt); delta < -time.Second || delta > time.Second {
			t.Fatalf("unexpected expiresAt delta: %v", delta)
		}
	})
}

func TestParseError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected int
	}{
		{name: "malformed", err: jwt.ErrTokenMalformed, expected: consts.ErrnoAuthJWTMalformedToken},
		{name: "signature invalid", err: jwt.ErrTokenSignatureInvalid, expected: consts.ErrnoAuthJWTSignatureInvalid},
		{name: "expired", err: jwt.ErrTokenExpired, expected: consts.ErrnoAuthJWTExpiredToken},
		{name: "unknown", err: stderrors.New("other"), expected: consts.ErrnoAuthJWTParseFailed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := parseError(tc.err)
			assertErrno(t, err, tc.expected)
		})
	}
}

func signToken(t *testing.T, secret []byte, claims customClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	return token
}

func signTokenWithMethod(t *testing.T, secret []byte, claims customClaims, method jwt.SigningMethod) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	return token
}

func assertErrno(t *testing.T, err error, code int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := commonerrors.Errno(err); got != code {
		t.Fatalf("unexpected errno: got %d, want %d, err=%v", got, code, err)
	}
	msg := consts.ErrMsg[code]
	if msg != "" && !strings.Contains(err.Error(), msg) {
		t.Fatalf("error message %q does not contain %q", err.Error(), msg)
	}
}
