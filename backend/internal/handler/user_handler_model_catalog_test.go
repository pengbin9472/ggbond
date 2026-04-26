package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pengbin9472/ggbond/internal/service"
	"github.com/stretchr/testify/require"
)

type stubAccountCatalogService struct {
	accounts []service.Account
}

func (s *stubAccountCatalogService) ListAccounts(_ context.Context, page, pageSize int, platform, accountType, status, search string, groupID int64, privacyMode, sortBy, sortOrder string) ([]service.Account, int64, error) {
	start := (page - 1) * pageSize
	if start >= len(s.accounts) {
		return nil, int64(len(s.accounts)), nil
	}
	end := start + pageSize
	if end > len(s.accounts) {
		end = len(s.accounts)
	}
	return s.accounts[start:end], int64(len(s.accounts)), nil
}

func setupUserModelCatalogRouter(accounts []service.Account) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewUserHandler(nil, &stubAccountCatalogService{accounts: accounts}, service.NewBillingService(nil, nil), nil, nil, nil, nil)
	router.GET("/api/v1/models/catalog", handler.GetModelCatalog)
	return router
}

func TestUserHandlerGetModelCatalog_HidesInternalIDs(t *testing.T) {
	router := setupUserModelCatalogRouter([]service.Account{
		{
			ID:       1,
			Name:     "acc-1",
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			GroupIDs: []int64{10},
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-haiku-4-5-20251001": "claude-haiku-4-5-20251001",
				},
			},
		},
		{
			ID:       2,
			Name:     "acc-2",
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			GroupIDs: []int64{11},
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-haiku-4-5-20251001": "claude-haiku-4-5-20251001",
				},
			},
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/catalog", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotContains(t, rec.Body.String(), "account_ids")
	require.NotContains(t, rec.Body.String(), "group_ids")

	var resp struct {
		Data struct {
			Models []modelCatalogEntry `json:"models"`
			Total  int                 `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Data.Total)
	require.Len(t, resp.Data.Models, 1)
	require.Equal(t, 2, resp.Data.Models[0].AccountCount)
	require.Equal(t, 2, resp.Data.Models[0].GroupCount)
	require.NotNil(t, resp.Data.Models[0].CacheWritePrice)
	require.NotNil(t, resp.Data.Models[0].CacheReadPrice)
}

func TestUserHandlerGetModelCatalog_IncludesImageOutputPricing(t *testing.T) {
	router := setupUserModelCatalogRouter([]service.Account{
		{
			ID:       1,
			Name:     "acc-image",
			Platform: service.PlatformGemini,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			GroupIDs: []int64{10},
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gemini-3-pro-image-preview": "gemini-3-pro-image-preview",
				},
			},
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/catalog", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Models []modelCatalogEntry `json:"models"`
			Total  int                 `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Data.Total)
	require.NotNil(t, resp.Data.Models[0].ImageOutputPrice)
}
