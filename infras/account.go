package infras

type CreateAccountDTO struct {
	Email      *string `json:"email"`
	Username   *string `json:"username"`
	Password   *string `json:"password"`
	IsBlock    *bool   `json:"is_block"`
}