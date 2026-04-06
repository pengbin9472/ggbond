package handler

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pengbin9472/ggbond/internal/handler/dto"
	"github.com/pengbin9472/ggbond/internal/pkg/antigravity"
	"github.com/pengbin9472/ggbond/internal/pkg/claude"
	"github.com/pengbin9472/ggbond/internal/pkg/geminicli"
	"github.com/pengbin9472/ggbond/internal/pkg/openai"
	"github.com/pengbin9472/ggbond/internal/pkg/response"
	middleware2 "github.com/pengbin9472/ggbond/internal/server/middleware"
	"github.com/pengbin9472/ggbond/internal/service"

	"github.com/gin-gonic/gin"
)

type accountCatalogService interface {
	ListAccounts(ctx context.Context, page, pageSize int, platform, accountType, status, search string, groupID int64, privacyMode string) ([]service.Account, int64, error)
}

// UserHandler handles user-related requests
type UserHandler struct {
	userService           *service.UserService
	accountCatalogService accountCatalogService
	billingService        *service.BillingService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *service.UserService, accountCatalogService accountCatalogService, billingService *service.BillingService) *UserHandler {
	return &UserHandler{
		userService:           userService,
		accountCatalogService: accountCatalogService,
		billingService:        billingService,
	}
}

type modelCatalogEntry struct {
	ID               string   `json:"id"`
	DisplayName      string   `json:"display_name"`
	Type             string   `json:"type"`
	Platform         string   `json:"platform"`
	InputPrice       *float64 `json:"input_price,omitempty"`
	OutputPrice      *float64 `json:"output_price,omitempty"`
	CacheWritePrice  *float64 `json:"cache_write_price,omitempty"`
	CacheReadPrice   *float64 `json:"cache_read_price,omitempty"`
	ImageOutputPrice *float64 `json:"image_output_price,omitempty"`
	AccountCount     int      `json:"account_count"`
	GroupCount       int      `json:"group_count"`
	PricingFallback  bool     `json:"pricing_fallback"`
}

type modelCatalogResponse struct {
	Models []modelCatalogEntry `json:"models"`
	Total  int                 `json:"total"`
}

type modelCatalogAccumulator struct {
	entry      modelCatalogEntry
	accountSet map[int64]struct{}
	groupSet   map[int64]struct{}
}

// ChangePasswordRequest represents the change password request payload
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest represents the update profile request payload
type UpdateProfileRequest struct {
	Username *string `json:"username"`
}

// GetProfile handles getting user profile
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userData, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserFromService(userData))
}

// ChangePassword handles changing user password
// POST /api/v1/users/me/password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svcReq := service.ChangePasswordRequest{
		CurrentPassword: req.OldPassword,
		NewPassword:     req.NewPassword,
	}
	err := h.userService.ChangePassword(c.Request.Context(), subject.UserID, svcReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Password changed successfully"})
}

// UpdateProfile handles updating user profile
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svcReq := service.UpdateProfileRequest{
		Username: req.Username,
	}
	updatedUser, err := h.userService.UpdateProfile(c.Request.Context(), subject.UserID, svcReq)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserFromService(updatedUser))
}

// GetModelCatalog returns the platform-wide model catalog for authenticated users.
// GET /api/v1/models/catalog
func (h *UserHandler) GetModelCatalog(c *gin.Context) {
	if h.accountCatalogService == nil {
		response.Error(c, 500, "Model catalog unavailable")
		return
	}

	const pageSize = 500

	accumulators := make(map[string]*modelCatalogAccumulator)
	page := 1

	for {
		accounts, total, err := h.accountCatalogService.ListAccounts(c.Request.Context(), page, pageSize, "", "", "", "", 0, "")
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}

		for i := range accounts {
			account := &accounts[i]
			for _, modelID := range accountAvailableModelIDs(account) {
				h.mergeModelCatalogEntry(accumulators, account, modelID)
			}
		}

		if len(accounts) < pageSize || int64(page*pageSize) >= total {
			break
		}
		page++
	}

	models := make([]modelCatalogEntry, 0, len(accumulators))
	for _, acc := range accumulators {
		acc.entry.AccountCount = len(acc.accountSet)
		acc.entry.GroupCount = len(acc.groupSet)
		models = append(models, acc.entry)
	}

	sort.Slice(models, func(i, j int) bool {
		if models[i].Platform != models[j].Platform {
			return models[i].Platform < models[j].Platform
		}
		return models[i].ID < models[j].ID
	})

	response.Success(c, modelCatalogResponse{
		Models: models,
		Total:  len(models),
	})
}

func accountAvailableModelIDs(account *service.Account) []string {
	if account == nil {
		return nil
	}

	if account.IsOpenAI() {
		if account.IsOpenAIPassthroughEnabled() {
			return openAIModelIDs(openai.DefaultModels)
		}
		if models := whitelistOrMappedModels(account); len(models) > 0 {
			return models
		}
		return openAIModelIDs(openai.DefaultModels)
	}

	if account.IsGemini() {
		if !account.IsOAuth() {
			if models := whitelistOrMappedModels(account); len(models) > 0 {
				return models
			}
		}
		return geminiModelIDs(geminicli.DefaultModels)
	}

	if account.Platform == service.PlatformAntigravity {
		if models := whitelistOrMappedModels(account); len(models) > 0 {
			return models
		}
		return antigravityModelIDs(antigravity.DefaultModels())
	}

	if account.Platform == service.PlatformSora {
		if models := whitelistOrMappedModels(account); len(models) > 0 {
			return models
		}
		return soraModelIDs(service.DefaultSoraModels(nil))
	}

	if account.IsOAuth() {
		if models := whitelistOrMappedModels(account); len(models) > 0 {
			return models
		}
		return claudeModelIDs(claude.DefaultModels)
	}

	if models := whitelistOrMappedModels(account); len(models) > 0 {
		return models
	}
	return claudeModelIDs(claude.DefaultModels)
}

func (h *UserHandler) mergeModelCatalogEntry(accumulators map[string]*modelCatalogAccumulator, account *service.Account, modelID string) {
	if strings.TrimSpace(modelID) == "" || account == nil {
		return
	}

	item, ok := accumulators[modelID]
	if !ok {
		item = &modelCatalogAccumulator{
			entry: modelCatalogEntry{
				ID:          modelID,
				DisplayName: modelID,
				Type:        "model",
				Platform:    inferModelPlatform(modelID, account.Platform),
			},
			accountSet: make(map[int64]struct{}),
			groupSet:   make(map[int64]struct{}),
		}
		if pricing, err := h.lookupModelPricing(modelID); err == nil && pricing != nil {
			item.entry.InputPrice = floatPtr(pricing.InputPricePerToken * 1_000_000)
			item.entry.OutputPrice = floatPtr(pricing.OutputPricePerToken * 1_000_000)
			item.entry.CacheWritePrice = floatPtr(pricing.CacheCreationPricePerToken * 1_000_000)
			item.entry.CacheReadPrice = floatPtr(pricing.CacheReadPricePerToken * 1_000_000)
			item.entry.ImageOutputPrice = floatPtr(pricing.ImageOutputPricePerToken * 1_000_000)
		} else {
			item.entry.PricingFallback = true
		}
		accumulators[modelID] = item
	}

	item.accountSet[account.ID] = struct{}{}
	for _, groupID := range account.GroupIDs {
		if groupID > 0 {
			item.groupSet[groupID] = struct{}{}
		}
	}
}

func (h *UserHandler) lookupModelPricing(modelID string) (*service.ModelPricing, error) {
	if h.billingService == nil {
		return nil, fmt.Errorf("billing service unavailable")
	}
	return h.billingService.GetModelPricing(modelID)
}

func whitelistOrMappedModels(account *service.Account) []string {
	if account == nil {
		return nil
	}

	whitelist := explicitWhitelistModels(account)
	if len(whitelist) > 0 {
		return whitelist
	}

	return mappedTargetModels(account)
}

func explicitWhitelistModels(account *service.Account) []string {
	result := make(map[string]struct{})

	if account != nil && account.Credentials != nil {
		if raw, ok := account.Credentials["model_whitelist"]; ok {
			switch v := raw.(type) {
			case []string:
				for _, model := range v {
					model = strings.TrimSpace(model)
					if model == "" || strings.Contains(model, "*") {
						continue
					}
					result[model] = struct{}{}
				}
			case []any:
				for _, item := range v {
					model, ok := item.(string)
					if !ok {
						continue
					}
					model = strings.TrimSpace(model)
					if model == "" || strings.Contains(model, "*") {
						continue
					}
					result[model] = struct{}{}
				}
			}
		}
	}

	for from, to := range account.GetModelMapping() {
		from = strings.TrimSpace(from)
		to = strings.TrimSpace(to)
		if from == "" || to == "" {
			continue
		}
		if strings.Contains(from, "*") || strings.Contains(to, "*") {
			continue
		}
		if from == to {
			result[from] = struct{}{}
		}
	}

	return sortedStringKeys(result)
}

func mappedTargetModels(account *service.Account) []string {
	result := make(map[string]struct{})
	for _, to := range account.GetModelMapping() {
		to = strings.TrimSpace(to)
		if to == "" || strings.Contains(to, "*") {
			continue
		}
		result[to] = struct{}{}
	}
	return sortedStringKeys(result)
}

func sortedStringKeys(values map[string]struct{}) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func claudeModelIDs(models []claude.Model) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	return ids
}

func geminiModelIDs(models []geminicli.Model) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	return ids
}

func openAIModelIDs(models []openai.Model) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	return ids
}

func antigravityModelIDs(models []antigravity.ClaudeModel) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	return ids
}

func soraModelIDs(models []openai.Model) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	return ids
}

func inferModelPlatform(modelID string, fallback string) string {
	lower := strings.ToLower(strings.TrimSpace(modelID))
	switch {
	case strings.HasPrefix(lower, "claude"):
		return service.PlatformAnthropic
	case strings.HasPrefix(lower, "gpt"), strings.HasPrefix(lower, "o1"), strings.HasPrefix(lower, "o3"), strings.HasPrefix(lower, "o4"), strings.HasPrefix(lower, "chatgpt"):
		return service.PlatformOpenAI
	case strings.HasPrefix(lower, "gemini"):
		return service.PlatformGemini
	case strings.HasPrefix(lower, "sora"), strings.HasPrefix(lower, "gpt-image"), strings.HasPrefix(lower, "prompt-enhance"):
		return service.PlatformSora
	default:
		return fallback
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
