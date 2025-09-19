package infras

type CreateGroupDTO struct {
	GroupName *string `json:"group_name"`
	GroupId   *string `json:"group_id"`
	AccountId *int32  `json:"account_id"`
}

type GetGroupsByAccountIDDTO struct {
	AccountID int32 `query:"account_id"`
}

type DeleteGroupDTO struct {
	GroupID int32 `json:"group_id"`
}