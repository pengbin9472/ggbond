package service

import (
	"github.com/pengbin9472/ggbond/internal/config"
	"github.com/pengbin9472/ggbond/internal/pkg/openai"
)

var defaultSoraModelIDs = []string{
	"gpt-image",
	"gpt-image-landscape",
	"gpt-image-portrait",
	"sora2-landscape-10s",
	"sora2-portrait-10s",
	"sora2-landscape-15s",
	"sora2-portrait-15s",
	"prompt-enhance-short-10s",
	"prompt-enhance-medium-10s",
}

// DefaultSoraModels keeps a lightweight compatibility surface for handlers that still
// need to expose legacy Sora model IDs during the merge transition.
func DefaultSoraModels(cfg *config.Config) []openai.Model {
	models := make([]openai.Model, 0, len(defaultSoraModelIDs))
	for _, id := range defaultSoraModelIDs {
		models = append(models, openai.Model{
			ID:          id,
			Object:      "model",
			OwnedBy:     "openai",
			Type:        "model",
			DisplayName: id,
		})
	}
	return models
}
