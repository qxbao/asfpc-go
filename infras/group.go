package infras

type CreateGroupDTO struct {
	GroupName *string `json:"group_name"`
	GroupId   *string `json:"group_id"`
	AccountId *int32  `json:"account_id"`
}