package ports

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainauth "github.com/leadtek-test/q1/container/domain/auth"
	pmiddleware "github.com/leadtek-test/q1/container/ports/middleware"
)

type fakeServer struct{}

func (f fakeServer) Register(c *gin.Context)              { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) Login(c *gin.Context)                 { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) Upload(c *gin.Context)                { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) CreateContainer(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) GetCreateContainerJob(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) ListContainers(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) UpdateContainerStatus(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
func (f fakeServer) DeleteContainer(c *gin.Context)       { c.JSON(http.StatusOK, gin.H{"ok": true}) }

type passTokenManager struct{}

func (p passTokenManager) Generate(uint, string) (string, time.Time, error) {
	return "", time.Time{}, nil
}
func (p passTokenManager) Parse(string) (domainauth.Claims, error) {
	return domainauth.Claims{UserID: 1, Username: "u"}, nil
}

func TestRegisterHandlersWithOption(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterHandlersWithOption(r, fakeServer{}, ServerOptions{
		BaseURL: "/api",
		ProtectedMiddlewares: []gin.HandlerFunc{
			pmiddleware.NewAuthMiddleware(passTokenManager{}).VerifyToken(),
		},
	})

	health := httptest.NewRecorder()
	r.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("healthz status: %d", health.Code)
	}

	protectedNoAuth := httptest.NewRecorder()
	r.ServeHTTP(protectedNoAuth, httptest.NewRequest(http.MethodGet, "/api/v1/containers", nil))
	if protectedNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("protected no auth status: %d", protectedNoAuth.Code)
	}

	protectedAuth := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/containers", nil)
	req.Header.Set("Authorization", "Bearer token")
	r.ServeHTTP(protectedAuth, req)
	if protectedAuth.Code != http.StatusOK {
		t.Fatalf("protected auth status: %d", protectedAuth.Code)
	}
}
