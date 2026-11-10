package muumuu

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Domain struct {
	ID                    string             `json:"id"`
	SLD                   string             `json:"sld"`
	TLD                   string             `json:"tld"`
	FQDN                  string             `json:"fqdn"`
	State                 string             `json:"state"`
	SetupState            string             `json:"setup-state"`
	Registrar             string             `json:"registrar,omitempty"`
	WhoisProxyEnabled     *bool              `json:"whois-proxy-enabled,omitempty"`
	AutoRenewEnabled      *bool              `json:"auto-renew-enabled,omitempty"`
	IsJapaneseDomain      *bool              `json:"is-japanese-domain,omitempty"`
	Contract              *DomainContract    `json:"contract,omitempty"`
	StatusGroup           string             `json:"status-group,omitempty"`
	ContractEndDate       *string            `json:"contract-end-date,omitempty"`
	ContractTotalTerm     *int               `json:"contract-total-term,omitempty"`
	PaymentDueDate        *string            `json:"payment-due-date,omitempty"`
	CancellationScheduled *bool              `json:"cancellation-scheduled,omitempty"`
	Expired               *DomainExpiredInfo `json:"expired,omitempty"`
	ContractState         *string            `json:"contract-state,omitempty"`
}

type DomainContract struct {
	ID               string     `json:"id"`
	State            string     `json:"state"`
	Term             *int       `json:"term,omitempty"`
	StartDate        *time.Time `json:"start-date,omitempty"`
	EndDate          *time.Time `json:"end-date,omitempty"`
	PaymentMethod    *string    `json:"payment-method,omitempty"`
	AutoRenewEnabled *bool      `json:"auto-renew-enabled,omitempty"`
}

type DomainExpiredInfo struct {
	DiscardedAt      *string `json:"discarded-at,omitempty"`
	RecoverableUntil string  `json:"recoverable-until"`
	RecoveryDays     int     `json:"recovery-days"`
	Updating         bool    `json:"updating"`
}

type InputGetDomains struct {
	ListOptions
	FQDN        string
	State       string
	StatusGroup []string
	Sort        string
	Q           string
}

type OutputGetDomains struct {
	Domains []Domain
	Meta    PaginationMeta
}

func (c *Client) GetDomains(ctx context.Context, input *InputGetDomains) (*OutputGetDomains, error) {
	q := url.Values{}
	if input != nil {
		input.setTo(q)
		setString(q, "fqdn", input.FQDN)
		setString(q, "state", input.State)
		setString(q, "status-group", strings.Join(input.StatusGroup, ","))
		setString(q, "sort", input.Sort)
		setString(q, "q", input.Q)
	}

	r, err := call[[]Domain](ctx, c, http.MethodGet, "/me/domains", q, nil)
	if err != nil {
		return nil, err
	}
	return &OutputGetDomains{Domains: r.Data, Meta: derefMeta(r.Meta)}, nil
}

func (c *Client) GetDomain(ctx context.Context, domainID string) (*Domain, error) {
	r, err := call[Domain](ctx, c, http.MethodGet, domainPath(domainID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

type InputUpdateDomain struct {
	AutoRenewEnabled *bool `json:"auto-renew-enabled,omitempty"`
}

func (c *Client) UpdateDomain(ctx context.Context, domainID string, input *InputUpdateDomain) (*Domain, error) {
	r, err := call[Domain](ctx, c, http.MethodPatch, domainPath(domainID), nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func domainPath(domainID string) string {
	return "/me/domains/" + url.PathEscape(domainID)
}

func derefMeta(m *PaginationMeta) PaginationMeta {
	if m == nil {
		return PaginationMeta{}
	}
	return *m
}
