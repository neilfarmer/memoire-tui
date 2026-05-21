package api

import "net/url"

type Favorite struct {
	FavoriteID  string   `json:"favorite_id"`
	Type        string   `json:"type,omitempty"`
	ResourceID  string   `json:"resource_id,omitempty"`
	URL         string   `json:"url,omitempty"`
	Title       string   `json:"title,omitempty"`
	FeedTitle   string   `json:"feed_title,omitempty"`
	Image       string   `json:"image,omitempty"`
	Description string   `json:"description,omitempty"`
	Published   string   `json:"published,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

type FavoriteInput struct {
	Type        string   `json:"type,omitempty"`
	ResourceID  string   `json:"resource_id,omitempty"`
	URL         string   `json:"url,omitempty"`
	Title       string   `json:"title,omitempty"`
	FeedTitle   string   `json:"feed_title,omitempty"`
	Image       string   `json:"image,omitempty"`
	Description string   `json:"description,omitempty"`
	Published   string   `json:"published,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (c *Client) ListFavorites() ([]Favorite, error) { return GetList[Favorite](c, "/favorites") }
func (c *Client) CreateFavorite(in FavoriteInput) (Favorite, error) {
	return PostOne[Favorite](c, "/favorites", in)
}
func (c *Client) UpdateFavoriteTags(id string, tags []string) (Favorite, error) {
	return PatchOne[Favorite](c, "/favorites/"+url.PathEscape(id), map[string]any{"tags": tags})
}
func (c *Client) DeleteFavorite(id string) error {
	return c.Delete("/favorites/" + url.PathEscape(id))
}
