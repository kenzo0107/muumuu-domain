package muumuu

import (
	"context"
	"net/http"
)

const (
	NameserverSetupTypeLolipop        = "lolipop"
	NameserverSetupTypeCustom         = "custom"
	NameserverSetupTypeAcquiredDomain = "acquired_domain"
	NameserverSetupTypeMuumuuDNS      = "muumuu_dns"
	NameserverSetupTypeParking        = "parking"
	NameserverSetupTypeOther          = "other"
)

type NameserverSettings struct {
	DomainID    string   `json:"domain-id"`
	SetupType   string   `json:"setup-type"`
	Nameservers []string `json:"nameservers"`
}

func (c *Client) GetNameservers(ctx context.Context, domainID string) (*NameserverSettings, error) {
	r, err := call[NameserverSettings](ctx, c, http.MethodGet, nameserversPath(domainID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

type InputUpdateNameservers struct {
	SetupType              string   `json:"setup-type"`
	Nameservers            []string `json:"nameservers,omitempty"`
	ConfirmExistingRecords *bool    `json:"confirm-existing-records,omitempty"`
}

func (c *Client) UpdateNameservers(ctx context.Context, domainID string, input *InputUpdateNameservers) (*NameserverSettings, error) {
	r, err := call[NameserverSettings](ctx, c, http.MethodPatch, nameserversPath(domainID), nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func nameserversPath(domainID string) string {
	return domainPath(domainID) + "/nameservers"
}
