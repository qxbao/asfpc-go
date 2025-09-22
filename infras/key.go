package infras

type AddGeminiKeyDTO struct {
	APIKey string `json:"api_key" binding:"required"`
}