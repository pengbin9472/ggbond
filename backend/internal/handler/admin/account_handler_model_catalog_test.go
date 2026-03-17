package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pengbin9472/ggbond/internal/service"
	"github.com/stretchr/testify/require"
)

func setupModelCatalogRouter(adminSvc service.AdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, service.NewBillingService(nil, nil), nil, nil, nil, nil, nil)
	router.GET("/api/v1/admin/accounts/models/catalog", handler.GetModelCatalog)
	return router
}

func TestAccountHandlerGetModelCatalog_DeduplicatesAcrossAccountsAndAddsPricing(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.accounts = []service.Account{
		{
			ID:       1,
			Name:     "acc-1",
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			GroupIDs: []int64{10},
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-haiku-4-5-20251001":  "claude-haiku-4-5-20251001",
					"claude-sonnet-4-5-20250929": "claude-sonnet-4-5-20250929",
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
	}

	router := setupModelCatalogRouter(adminSvc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/models/catalog", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Models []adminModelCatalogEntry `json:"models"`
			Total  int                      `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 2, resp.Data.Total)

	var haiku *adminModelCatalogEntry
	for i := range resp.Data.Models {
		if resp.Data.Models[i].ID == "claude-haiku-4-5-20251001" {
			haiku = &resp.Data.Models[i]
			break
		}
	}
	require.NotNil(t, haiku)
	require.Equal(t, 2, haiku.AccountCount)
	require.Equal(t, 2, haiku.GroupCount)
	require.NotNil(t, haiku.InputPrice)
	require.NotNil(t, haiku.OutputPrice)
}

func TestAccountHandlerGetModelCatalog_UsesMappingTargetWhenWhitelistMissing(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.accounts = []service.Account{
		{
			ID:       1,
			Name:     "acc-map",
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			GroupIDs: []int64{12},
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"claude-opus-*": "claude-opus-4-5-20251101",
				},
			},
		},
	}

	router := setupModelCatalogRouter(adminSvc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/models/catalog", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Models []adminModelCatalogEntry `json:"models"`
			Total  int                      `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Data.Total)
	require.Equal(t, "claude-opus-4-5-20251101", resp.Data.Models[0].ID)
}
