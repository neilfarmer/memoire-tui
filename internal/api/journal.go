package api

import "net/url"

type JournalSummary struct {
	EntryDate    string `json:"entry_date"`
	TitlePreview string `json:"title_preview,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type JournalEntry struct {
	EntryDate string `json:"entry_date"`
	Title     string `json:"title,omitempty"`
	// The API has historically returned the entry text under either
	// "content" or "body". Keep both DTO fields and use Text() to read.
	Content   string   `json:"content,omitempty"`
	Body      string   `json:"body,omitempty"`
	Mood      string   `json:"mood,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

// Text returns the entry body, preferring Content over Body when both
// are populated.
func (e JournalEntry) Text() string {
	if e.Content != "" {
		return e.Content
	}
	return e.Body
}

type JournalInput struct {
	Title   string   `json:"title,omitempty"`
	Content string   `json:"content,omitempty"`
	Body    string   `json:"body,omitempty"`
	Mood    string   `json:"mood,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

func (c *Client) ListJournal(query string) ([]JournalSummary, error) {
	q := url.Values{}
	if query != "" {
		q.Set("q", query)
	}
	var out []JournalSummary
	return out, c.GetQ("/journal", q, &out)
}

func (c *Client) GetJournal(date string) (JournalEntry, error) {
	return GetOne[JournalEntry](c, "/journal/"+url.PathEscape(date))
}
func (c *Client) UpsertJournal(date string, in JournalInput) (JournalEntry, error) {
	return PutOne[JournalEntry](c, "/journal/"+url.PathEscape(date), in)
}
func (c *Client) DeleteJournal(date string) error {
	return c.Delete("/journal/" + url.PathEscape(date))
}
