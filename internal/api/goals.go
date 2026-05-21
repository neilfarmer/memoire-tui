package api

import "net/url"

type Goal struct {
	GoalID      string `json:"goal_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Status      string `json:"status,omitempty"`
	Deadline    string `json:"deadline,omitempty"`
	TargetDate  string `json:"target_date,omitempty"`
	Progress    int    `json:"progress,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type GoalInput struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Status      string `json:"status,omitempty"`
	Deadline    string `json:"deadline,omitempty"`
	TargetDate  string `json:"target_date,omitempty"`
	Progress    int    `json:"progress,omitempty"`
}

func (c *Client) ListGoals() ([]Goal, error) { return GetList[Goal](c, "/goals") }
func (c *Client) GetGoal(id string) (Goal, error) {
	return GetOne[Goal](c, "/goals/"+url.PathEscape(id))
}
func (c *Client) CreateGoal(in GoalInput) (Goal, error) { return PostOne[Goal](c, "/goals", in) }
func (c *Client) UpdateGoal(id string, in GoalInput) (Goal, error) {
	return PutOne[Goal](c, "/goals/"+url.PathEscape(id), in)
}
func (c *Client) DeleteGoal(id string) error { return c.Delete("/goals/" + url.PathEscape(id)) }
