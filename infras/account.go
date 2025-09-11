package infras

type CreateAccountDTO struct {
	Email      *string `json:"email"`
	Username   *string `json:"username"`
	Password   *string `json:"password"`
	IsBlock    *bool   `json:"is_block"`
}

type GetAccountsDTO struct {
	Page  *int32 `query:"page"`
	Limit *int32 `query:"limit"`
}