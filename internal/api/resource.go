package api

// Generic helpers for the standard CRUD verbs used by per-resource files.
// Each per-resource file exposes its public *Client methods as one-liners
// over these so the JSON-decode boilerplate lives in one place.

// GetList GETs a JSON array and decodes it into []T.
func GetList[T any](c *Client, path string) ([]T, error) {
	var out []T
	return out, c.Get(path, &out)
}

// GetOne GETs a single object and decodes it into T.
func GetOne[T any](c *Client, path string) (T, error) {
	var out T
	return out, c.Get(path, &out)
}

// PostOne POSTs body and decodes the response into T.
func PostOne[T any](c *Client, path string, body any) (T, error) {
	var out T
	return out, c.Post(path, body, &out)
}

// PutOne PUTs body and decodes the response into T.
func PutOne[T any](c *Client, path string, body any) (T, error) {
	var out T
	return out, c.Put(path, body, &out)
}

// PatchOne PATCHes body and decodes the response into T.
func PatchOne[T any](c *Client, path string, body any) (T, error) {
	var out T
	return out, c.Patch(path, body, &out)
}
