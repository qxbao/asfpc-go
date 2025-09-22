package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type PromptService struct {
	Server infras.Server
}

func (s *PromptService) GetPrompt(ctx context.Context, promptName string) (db.Prompt, error) {
	prompt, err := s.Server.Queries.GetPrompt(ctx, promptName)
	if err != nil {
		return db.Prompt{}, fmt.Errorf("%v", err)
	}

	return prompt, nil
}

func (s *PromptService) ReplacePrompt(prompt *string, kwargs ...string) string {
	for i, kw := range kwargs {
		placeholder := fmt.Sprintf("INSERT_%d", i)
		replacement := kw
		if kw == "" {
			replacement = "NULL"
		}
		*prompt = strings.ReplaceAll(*prompt, placeholder, replacement)
	}
	return *prompt
}