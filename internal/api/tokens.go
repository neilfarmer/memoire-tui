package api

import "net/url"

type Token struct {
	TokenID    string `json:"token_id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at,omitempty"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	Token      string `json:"token,omitempty"`
}

func (c *Client) ListTokens() ([]Token, error) { return GetList[Token](c, "/tokens") }
func (c *Client) CreateToken(name string) (Token, error) {
	return PostOne[Token](c, "/tokens", map[string]string{"name": name})
}
func (c *Client) DeleteToken(id string) error { return c.Delete("/tokens/" + url.PathEscape(id)) }
