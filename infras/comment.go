package infras

type GetPostCommentsResponse struct {
	Data   *[]Comment `json:"data"`
	Paging *Paging    `json:"paging,omitempty"`
}

type Comment struct {
	ID          *string `json:"id"`
	CreatedTime *string `json:"created_time"`
	Message     *string `json:"message,omitempty"`
	From        *Author `json:"from"`
}

type Author struct {
	Name *string `json:"name"`
	ID   *string `json:"id"`
}