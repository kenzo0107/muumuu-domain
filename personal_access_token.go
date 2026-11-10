package muumuu

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type PersonalAccessToken struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token-prefix"`
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires-at"`
	LastUsedAt  *time.Time `json:"last-used-at"`
	Status      string     `json:"status"`
}

func (c *Client) GetPersonalAccessTokens(ctx context.Context) ([]PersonalAccessToken, error) {
	r, err := call[[]PersonalAccessToken](ctx, c, http.MethodGet, "/me/personal-access-tokens", nil, nil)
	if err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (c *Client) DeletePersonalAccessToken(ctx context.Context, id int64) error {
	return c.callNoContent(ctx, http.MethodDelete, "/me/personal-access-tokens/"+strconv.FormatInt(id, 10))
}
