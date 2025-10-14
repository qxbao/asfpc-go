package infras

type CreateModelRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	CategoryID  *int32 `json:"category_id"` // nullable
}

type UpdateModelRequest struct {
	ID          int32  `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	CategoryID  *int32 `json:"category_id"` // nullable
}

type DeleteModelRequest struct {
	ID int32 `json:"id" validate:"required"`
}

type AssignModelToCategoryRequest struct {
	ModelID    int32 `json:"model_id" validate:"required"`
	CategoryID int32 `json:"category_id" validate:"required"`
}
