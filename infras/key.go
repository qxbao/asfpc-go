package infras

type AddGeminiKeyDTO struct {
	APIKey string `json:"api_key" binding:"required"`
}

type DeleteGeminiKeyDTO struct {
	KeyID int32 `json:"key_id" binding:"required"`
}