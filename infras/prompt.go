package infras

type CreatePromptRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Content     string `json:"content" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
}
