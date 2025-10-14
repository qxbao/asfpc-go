package infras

type CreatePromptRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	CategoryID  int    `json:"category_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

type DeletePromptRequest struct {
	ID int32 `json:"id" binding:"required"`
}

type RollbackPromptRequest struct {
	CategoryID int32 `json:"category_id" binding:"required"`
	ServiceName string `json:"service_name" binding:"required"`
}