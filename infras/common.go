package infras

type Paging struct {
	Previous *string `json:"previous,omitempty"`
	Next     *string `json:"next,omitempty"`
}