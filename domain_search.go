package muumuu

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	DomainAvailabilityAvailable   = "available"
	DomainAvailabilityUnavailable = "unavailable"
	DomainAvailabilityPremium     = "premium"
)

type DomainSearchResult struct {
	FQDN              string  `json:"fqdn"`
	Availability      string  `json:"availability"`
	UnavailableReason *string `json:"unavailable-reason"`
	DomainPrice       *int    `json:"domain-price"`
	ServiceFee        *int    `json:"service-fee"`
	Tax               *int    `json:"tax"`
	Total             *int    `json:"total"`
	Currency          string  `json:"currency"`
}

type DomainSearchData struct {
	Query     string               `json:"query"`
	Results   []DomainSearchResult `json:"results"`
	Truncated bool                 `json:"truncated"`
}

type InputSearchDomains struct {
	Q    string
	TLDs []string
}

func (c *Client) SearchDomains(ctx context.Context, input *InputSearchDomains) (*DomainSearchData, error) {
	if input == nil || input.Q == "" {
		return nil, fmt.Errorf("q is required")
	}

	q := url.Values{}
	q.Set("q", input.Q)
	setString(q, "tlds", strings.Join(input.TLDs, ","))

	r, err := call[DomainSearchData](ctx, c, http.MethodGet, "/domain-search", q, nil)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}
