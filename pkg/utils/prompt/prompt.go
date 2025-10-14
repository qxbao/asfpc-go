package prompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type PromptService struct {
	Server *infras.Server
}

func (s *PromptService) GetPrompt(ctx context.Context, promptName string, categoryId int32) (db.Prompt, error) {
	prompt, err := s.Server.Queries.GetPrompt(ctx, db.GetPromptParams{
		ServiceName: promptName,
		CategoryID:  categoryId,
	})
	if err != nil {
		return db.Prompt{}, fmt.Errorf("%v", err)
	}

	return prompt, nil
}

func (s *PromptService) ReplacePrompt(prompt string, kwargs ...string) string {
	for i, kw := range kwargs {
		placeholder := fmt.Sprintf("INSERT_%d", (i + 1))
		replacement := kw
		if kw == "" {
			replacement = "(null)"
		}
		prompt = strings.Replace(prompt, placeholder, replacement, 1)
	}
	return prompt
}