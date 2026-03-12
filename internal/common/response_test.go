package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
)

func TestBaseResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	base := BaseResponse{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	base.Response(c, nil, gin.H{"x": 1})

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if int(resp["errno"].(float64)) != consts.ErrnoSuccess {
		t.Fatalf("unexpected errno: %v", resp["errno"])
	}
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	base.Response(c, commonerrors.New(consts.ErrnoAuthInvalidToken), nil)

	resp = map[string]any{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if int(resp["errno"].(float64)) != consts.ErrnoAuthInvalidToken {
		t.Fatalf("unexpected errno: %v", resp["errno"])
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
