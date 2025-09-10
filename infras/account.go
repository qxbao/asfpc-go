package infras

type AccountRequest struct {
	Username    *string `json:"username"`
	Password    *string `json:"password"`
	IsBlock     *bool   `json:"is_block"`
}