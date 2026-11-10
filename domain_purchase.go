package muumuu

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

const (
	PurchaseStatusPending    = "pending"
	PurchaseStatusProcessing = "processing"
	PurchaseStatusCompleted  = "completed"
	PurchaseStatusFailed     = "failed"
)

type InputQuoteDomainPurchase struct {
	FQDN string `json:"fqdn"`
	Term *int   `json:"term,omitempty"`
}

type PurchaseToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires-at"`
}

type DomainPurchaseQuote struct {
	FQDN                      string         `json:"fqdn"`
	Availability              string         `json:"availability"`
	Term                      int            `json:"term"`
	DomainPrice               *int           `json:"domain-price"`
	ServiceFee                *int           `json:"service-fee"`
	Tax                       *int           `json:"tax"`
	Total                     *int           `json:"total"`
	Currency                  string         `json:"currency"`
	CreditCardRegistered      bool           `json:"credit-card-registered"`
	CreditCardRegistrationURL *string        `json:"credit-card-registration-url"`
	PurchaseToken             *PurchaseToken `json:"purchase-token"`
}

func (c *Client) QuoteDomainPurchase(ctx context.Context, input *InputQuoteDomainPurchase) (*DomainPurchaseQuote, error) {
	r, err := call[DomainPurchaseQuote](ctx, c, http.MethodPost, "/me/domain-purchase/quote", nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

type InputPurchaseDomain struct {
	PurchaseToken string `json:"purchase-token"`
}

type DomainPurchase struct {
	PurchaseID  int64     `json:"purchase-id"`
	CheckoutURL string    `json:"checkout-url"`
	Status      string    `json:"status"`
	FQDN        string    `json:"fqdn"`
	Term        int       `json:"term"`
	ExpiresAt   time.Time `json:"expires-at"`
}

func (c *Client) PurchaseDomain(ctx context.Context, input *InputPurchaseDomain) (*DomainPurchase, error) {
	r, err := call[DomainPurchase](ctx, c, http.MethodPost, "/me/domain-purchase", nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

type DirectDomainPurchase struct {
	PurchaseID  string `json:"purchase-id"`
	Status      string `json:"status"`
	FQDN        string `json:"fqdn"`
	Term        int    `json:"term"`
	DomainID    string `json:"domain-id"`
	ContractID  string `json:"contract-id"`
	CompletedAt string `json:"completed-at"`
}

func (c *Client) PurchaseDomainDirect(ctx context.Context, input *InputPurchaseDomain) (*DirectDomainPurchase, error) {
	r, err := call[DirectDomainPurchase](ctx, c, http.MethodPost, "/me/domain-purchase/direct", nil, input)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}

type DomainPurchaseStatusDomain struct {
	DomainName    string  `json:"domain-name"`
	Term          int     `json:"term"`
	Status        string  `json:"status"`
	DomainID      *string `json:"domain-id"`
	ContractID    *string `json:"contract-id"`
	FailureReason *string `json:"failure-reason"`
}

type DomainPurchaseStatus struct {
	PurchaseID int64                        `json:"purchase-id"`
	Status     string                       `json:"status"`
	FQDN       string                       `json:"fqdn"`
	Term       int                          `json:"term"`
	Domains    []DomainPurchaseStatusDomain `json:"domains"`
}

func (c *Client) GetDomainPurchaseStatus(ctx context.Context, purchaseID int64) (*DomainPurchaseStatus, error) {
	path := "/me/domain-purchase/" + strconv.FormatInt(purchaseID, 10) + "/status"
	r, err := call[DomainPurchaseStatus](ctx, c, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}
