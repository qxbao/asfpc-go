package infras

type AccountRequest struct {
	Username    *string `json:"username"`
	AccessToken *string `json:"access_token"`
	IsBlock     *bool  `json:"is_block"`
}