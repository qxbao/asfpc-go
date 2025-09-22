package generative

import (
	"context"
	"fmt"

	"github.com/qxbao/asfpc/db"
	"google.golang.org/genai"
)

type GenerativeService struct {
	APIKey string
	Model  string
	context.Context
	client *genai.Client
	Usage  int64
}

func GetGenerativeService(apiKey, model string) *GenerativeService {
	return &GenerativeService{
		APIKey:  apiKey,
		Model:   model,
		Context: context.Background(),
		Usage:   0,
	}
}

func (gs *GenerativeService) Init() error {
	client, err := genai.NewClient(gs.Context, &genai.ClientConfig{
		APIKey:  gs.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	gs.client = client
	return err
}

func (gs *GenerativeService) GenerateText(prompt string) (string, error) {
	response, err := gs.client.Models.GenerateContent(gs.Context, gs.Model, genai.Text(prompt), nil)
	
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	gs.Usage += int64(response.UsageMetadata.PromptTokenCount)
	
	return response.Text(), nil
}

func (gs *GenerativeService) SaveUsage(ctx context.Context, queries *db.Queries) error {
	if gs.Usage > 0 {
		_, err := queries.UpdateGeminiKeyUsage(ctx, db.UpdateGeminiKeyUsageParams{
			ApiKey:    gs.APIKey,
			TokenUsed: gs.Usage,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}