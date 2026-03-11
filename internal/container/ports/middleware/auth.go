package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common"
	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/auth"
	"github.com/leadtek-test/q1/container/ports/contextx"
)

type AuthMiddleware struct {
	common.BaseResponse
	tokenManager auth.TokenManager
}

func NewAuthMiddleware(tokenManager auth.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenManager: tokenManager,
	}
}

func (m *AuthMiddleware) VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.tokenManager == nil {
			m.abortWithError(c, errors.New(consts.ErrnoAuthTokenManagerUnavailable))
			return
		}

		raw := c.GetHeader("Authorization")
		if raw == "" {
			m.abortWithError(c, errors.New(consts.ErrnoAuthMissingAuthorizationHeader))
			return
		}
		if !strings.HasPrefix(raw, "Bearer ") {
			m.abortWithError(c, errors.New(consts.ErrnoAuthInvalidAuthorizationHeader))
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		if token == "" {
			m.abortWithError(c, errors.New(consts.ErrnoAuthInvalidAuthorizationHeader))
			return
		}

		claims, err := m.tokenManager.Parse(token)
		if err != nil {
			m.abortWithError(c, err)
			return
		}

		c.Set(contextx.KeyUserID, claims.UserID)
		c.Set(contextx.KeyUsername, claims.Username)
		c.Next()
	}
}

func (m *AuthMiddleware) abortWithError(c *gin.Context, err error) {
	m.Response(c, err, nil)
	c.Abort()
}
