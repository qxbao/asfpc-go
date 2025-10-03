package infras

type Paging struct {
	Previous *string `json:"previous,omitempty"`
	Next     *string `json:"next,omitempty"`
}

type EntityNameID struct {
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name"`
}

type RoutingService struct {
	Server *Server
}