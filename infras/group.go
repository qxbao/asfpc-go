package infras

type CreateGroupDTO struct {
	GroupName *string `json:"group_name"`
	GroupId   *string `json:"group_id"`
	AccountId *int32  `json:"account_id"`
}

type GetGroupPostsResponse struct {
	Data   *[]Post  `json:"data"`
	Paging *Paging `json:"paging,omitempty"`
}

type Paging struct {
	Previous *string `json:"previous,omitempty"`
	Next     *string `json:"next,omitempty"`
}

type Post struct {
	ID          *string `json:"id"`
	UpdatedTime *string `json:"updated_time"`
	Message     *string `json:"message,omitempty"`
	IsBroadcast *bool   `json:"is_broadcast"`
}
