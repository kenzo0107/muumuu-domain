package muumuu

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DNSRecordTypeA     = "A"
	DNSRecordTypeAAAA  = "AAAA"
	DNSRecordTypeCNAME = "CNAME"
	DNSRecordTypeMX    = "MX"
	DNSRecordTypeTXT   = "TXT"
	DNSRecordTypeNS    = "NS"
	DNSRecordTypeALIAS = "ALIAS"
	DNSRecordTypeSRV   = "SRV"
	DNSRecordTypeCAA   = "CAA"
)

type DNSRecord struct {
	ID        int64      `json:"id"`
	FQDN      string     `json:"fqdn"`
	Type      string     `json:"type"`
	Value     string     `json:"value"`
	TTL       int        `json:"ttl"`
	Priority  *int       `json:"priority,omitempty"`
	CreatedAt *time.Time `json:"created-at,omitempty"`
	UpdatedAt *time.Time `json:"updated-at,omitempty"`
}

type InputGetDNSRecords struct {
	ListOptions
	Type string
	FQDN string
}

func (i *InputGetDNSRecords) values() url.Values {
	q := url.Values{}
	if i != nil {
		i.setTo(q)
		setString(q, "type", i.Type)
		setString(q, "fqdn", i.FQDN)
	}
	return q
}

type OutputGetDNSRecords struct {
	DNSRecords []DNSRecord
	Meta       PaginationMeta
}

func (c *Client) GetDNSRecords(ctx context.Context, domainID string, input *InputGetDNSRecords) (*OutputGetDNSRecords, error) {
	r, err := call[[]DNSRecord](ctx, c, http.MethodGet, dnsRecordsPath(domainID), input.values(), nil)
	if err != nil {
		return nil, err
	}
	return &OutputGetDNSRecords{DNSRecords: r.Data, Meta: derefMeta(r.Meta)}, nil
}

type InputGetDNSRecordsByDomainFQDN struct {
	InputGetDNSRecords
	DomainFQDN string
}

func (c *Client) GetDNSRecordsByDomainFQDN(ctx context.Context, input *InputGetDNSRecordsByDomainFQDN) (*OutputGetDNSRecords, error) {
	if input == nil || input.DomainFQDN == "" {
		return nil, fmt.Errorf("DomainFQDN is required")
	}

	q := input.values()
	q.Set("domain-fqdn", input.DomainFQDN)

	r, err := call[[]DNSRecord](ctx, c, http.MethodGet, "/me/dns-records", q, nil)
	if err != nil {
		return nil, err
	}
	return &OutputGetDNSRecords{DNSRecords: r.Data, Meta: derefMeta(r.Meta)}, nil
}

// Why not GET /dns-records/{record-id}: the API does not provide it.
func (c *Client) GetDNSRecord(ctx context.Context, domainID string, recordID int64) (*DNSRecord, error) {
	input := &InputGetDNSRecords{ListOptions: ListOptions{Page: 1, PageSize: 100}}
	for {
		out, err := c.GetDNSRecords(ctx, domainID, input)
		if err != nil {
			return nil, err
		}
		for i := range out.DNSRecords {
			if out.DNSRecords[i].ID == recordID {
				return &out.DNSRecords[i], nil
			}
		}
		if !out.Meta.HasNext() || len(out.DNSRecords) == 0 {
			return nil, &Error{
				StatusCode: http.StatusNotFound,
				Code:       ErrorCodeNotFound,
				Message:    fmt.Sprintf("dns record %d not found in domain %s", recordID, domainID),
			}
		}
		input.Page++
	}
}

type InputCreateDNSRecord struct {
	FQDN     string `json:"fqdn"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Priority *int   `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

func (c *Client) CreateDNSRecord(ctx context.Context, domainID string, input *InputCreateDNSRecord) (*DNSRecord, error) {
	r, err := call[DNSRecord](ctx, c, http.MethodPost, dnsRecordsPath(domainID), nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

type InputUpdateDNSRecord struct {
	Value    *string `json:"value,omitempty"`
	Priority *int    `json:"priority,omitempty"`
}

func (c *Client) UpdateDNSRecord(ctx context.Context, domainID string, recordID int64, input *InputUpdateDNSRecord) (*DNSRecord, error) {
	r, err := call[DNSRecord](ctx, c, http.MethodPatch, dnsRecordPath(domainID, recordID), nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, domainID string, recordID int64) error {
	return c.callNoContent(ctx, http.MethodDelete, dnsRecordPath(domainID, recordID))
}

func dnsRecordsPath(domainID string) string {
	return domainPath(domainID) + "/dns-records"
}

func dnsRecordPath(domainID string, recordID int64) string {
	return dnsRecordsPath(domainID) + "/" + strconv.FormatInt(recordID, 10)
}
