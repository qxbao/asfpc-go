package infras

type CreateAccountDTO struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	IsBlock  *bool   `json:"is_block"`
}

type QueryWithPageDTO struct {
	Page  *int32 `query:"page"`
	Limit *int32 `query:"limit"`
}

type GetAccountDTO struct {
	ID int32 `query:"id"`
}

type LoginAccountDTO struct {
	UID int32 `json:"uid"`
}

type JoinGroupDTO struct {
	GID int32 `json:"gid"`
}

type DeleteAccountsDTO struct {
	IDs []int32 `json:"ids"`
}

type GenAccountsATDTO struct {
	IDs []int32 `json:"ids"`
}

type UpdateAccountCredentialsDTO struct {
	ID       int32   `json:"id"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password *string `json:"password"`
}
