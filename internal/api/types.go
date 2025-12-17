package api

import "net/url"

func (c *Client) cloneBaseURL() *url.URL {
	clone := *c.baseURL
	return &clone
}
