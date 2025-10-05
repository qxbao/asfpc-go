package infras

import (
	"context"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/pkg/generative"
	"github.com/qxbao/asfpc/pkg/utils/python"
)

type GeminiScoringTaskInput struct {
	Ctx     context.Context
	Gs      *generative.GenerativeService
	Prompt  string
	Profile *db.GetProfilesAnalysisCronjobRow
}

type GeminiEmbeddingTaskInput struct {
	Ctx     context.Context
	Id      int32
	Ps      python.PythonService
}

type FindSimilarProfilesDTO struct {
	ProfileID *int32   `query:"profile_id" validate:"required"`
	TopK      *int32   `query:"top_k" validate:"min=1,max=20"`
}