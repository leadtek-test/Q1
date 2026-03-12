package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domainauth "github.com/leadtek-test/q1/container/domain/auth"
	"github.com/leadtek-test/q1/container/ports/contextx"
)

type fakeTokenManager struct {
	parseFn func(string) (domainauth.Claims, error)
}

func (f fakeTokenManager) Generate(uint, string) (string, time.Time, error) {
	return "", time.Time{}, nil
}
func (f fakeTokenManager) Parse(token string) (domainauth.Claims, error) {
	if f.parseFn != nil {
		return f.parseFn(token)
	}
	return domainauth.Claims{UserID: 1, Username: "u"}, nil
}

func TestAuthMiddlewareVerifyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing header", func(t *testing.T) {
		r := gin.New()
		r.Use(NewAuthMiddleware(fakeTokenManager{}).VerifyToken())
		r.GET("/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		r := gin.New()
		r.Use(NewAuthMiddleware(fakeTokenManager{
			parseFn: func(string) (domainauth.Claims, error) {
				return domainauth.Claims{}, commonerrors.New(consts.ErrnoAuthInvalidToken)
			},
		}).VerifyToken())
		r.GET("/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", "Bearer bad")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", w.Code)
		}
	})

	t.Run("success sets context", func(t *testing.T) {
		r := gin.New()
		r.Use(NewAuthMiddleware(fakeTokenManager{
			parseFn: func(string) (domainauth.Claims, error) {
				return domainauth.Claims{UserID: 7, Username: "alice"}, nil
			},
		}).VerifyToken())
		r.GET("/x", func(c *gin.Context) {
			if c.GetUint(contextx.KeyUserID) != 7 {
				t.Fatalf("user id not set")
			}
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", "Bearer good")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", w.Code)
		}
	})
}
