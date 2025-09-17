package infras

import (
	"encoding/json"
	"fmt"
)

type FlexibleID string

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleID(s)
		return nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleID(n.String())
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexibleID", data)
}

func (f FlexibleID) String() string {
	return string(f)
}

type GetGroupPostsResponse struct {
	Data   *[]Post `json:"data"`
	Paging *Paging `json:"paging,omitempty"`
}

type Post struct {
	ID          *string                  `json:"id"`
	UpdatedTime *string                  `json:"updated_time"`
	Message     *string                  `json:"message,omitempty"`
	IsBroadcast *bool                    `json:"is_broadcast"`
	Actions     *[]Action                `json:"actions,omitempty"`
	Comments    *CommentsData            `json:"comments,omitempty"`
	From        *FromUser                `json:"from,omitempty"`
	Icon        *string                  `json:"icon,omitempty"`
	IsHidden    *bool                    `json:"is_hidden,omitempty"`
	IsExpired   *bool                    `json:"is_expired,omitempty"`
	Link        *string                  `json:"link,omitempty"`
	Name        *string                  `json:"name,omitempty"`
	ObjectID    *string                  `json:"object_id,omitempty"`
	Picture     *string                  `json:"picture,omitempty"`
	Privacy     *Privacy                 `json:"privacy,omitempty"`
	Properties  *[]Property              `json:"properties,omitempty"`
	Source      *string                  `json:"source,omitempty"`
	StatusType  *string                  `json:"status_type,omitempty"`
	Subscribed  *bool                    `json:"subscribed,omitempty"`
	To          *ToData                  `json:"to,omitempty"`
	Type        *string                  `json:"type,omitempty"`
	CreatedTime *string                  `json:"created_time,omitempty"`
	MessageTags *map[string][]MessageTag `json:"message_tags,omitempty"`
}

type Action struct {
	Name *string `json:"name"`
	Link *string `json:"link"`
}

type CommentsData struct {
	Count *int           `json:"count"`
	Data  *[]PostComment `json:"data,omitempty"`
}

type PostComment struct {
	ID          *string        `json:"id"`
	Message     *string        `json:"message"`
	MessageTags *[]interface{} `json:"message_tags"`
	CreatedTime *string        `json:"created_time"`
	Likes       *int           `json:"likes"`
	From        *FromUser      `json:"from"`
}

type FromUser struct {
	ID           *FlexibleID `json:"id"`
	Name         *string     `json:"name"`
	Category     *string     `json:"category,omitempty"`
	CategoryList *[]Category `json:"category_list,omitempty"`
}

type Category struct {
	ID   *FlexibleID `json:"id"`
	Name *string     `json:"name"`
}

type Privacy struct {
	Allow       *string `json:"allow"`
	Deny        *string `json:"deny"`
	Description *string `json:"description"`
	Friends     *string `json:"friends"`
	Value       *string `json:"value"`
}

type Property struct {
	Name *string `json:"name"`
	Text *string `json:"text"`
}

type ToData struct {
	Data *[]ToItem `json:"data"`
}

type ToItem struct {
	ID   *FlexibleID `json:"id"`
	Name *string     `json:"name"`
}

type MessageTag struct {
	ID     *FlexibleID `json:"id"`
	Name   *string     `json:"name"`
	Offset *int        `json:"offset"`
	Length *int        `json:"length"`
}
