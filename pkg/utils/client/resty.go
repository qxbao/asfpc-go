package client

import "resty.dev/v3"

func NewRestyClient() *resty.Client {
	c := resty.New()
	c.SetRetryCount(3)
	return c
}