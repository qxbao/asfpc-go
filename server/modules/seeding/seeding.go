package seeding

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/qxbao/asfpc/db"
)

type Seed struct {
	Configs map[string]string `json:"config"`
	Prompts map[string]string `json:"prompt"`
}

func SeedData(logger zap.SugaredLogger, queries *db.Queries) {
	logger.Info("Generating seed data...")

	exe, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path:", err)
		return
	}

	data, err := os.ReadFile(path.Join(path.Dir(exe), "seed.json"))
	if err != nil {
		logger.Error("Failed to read seed.json:", err)
		return
	}

	var seed Seed
	if err := json.Unmarshal(data, &seed); err != nil {
		logger.Error("Failed to unmarshal seed.json:", err)
		return
	}

	ctx := context.Background()
	for key, value := range seed.Configs {
		_, err = queries.UpsertConfig(ctx, db.UpsertConfigParams{
			Key: key, Value: value,
		})
		if err != nil {
			logger.Error("Failed to upsert config:", err)
		}
	}

	categories, err := queries.GetCategories(ctx)
	if err != nil {
		logger.Error("Failed to get categories:", err)
		return
	}
	for name, content := range seed.Prompts {
		for _, c := range categories {
			_, err := queries.GetPrompt(ctx, db.GetPromptParams{
				ServiceName: name,
				CategoryID:  c.ID,
			})
			if err == nil {
				continue
			}
			_, _ = queries.CreatePrompt(ctx, db.CreatePromptParams{
				ServiceName: name,
				Content:     content,
				CreatedBy:   "system",
			})
		}
	}
	logger.Info("Seed data generated successfully")
}

var SeedModule = fx.Module(
	"seed",
	fx.Invoke(SeedData),
)
