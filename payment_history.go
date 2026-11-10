package muumuu

import (
	"context"
	"net/http"
	"net/url"
)

type PaymentHistoryAmountBreakdown struct {
	BasePrice int `json:"base-price"`
	Tax       int `json:"tax"`
	Surcharge int `json:"surcharge"`
	Discount  int `json:"discount"`
}

type PaymentHistory struct {
	ID                string                        `json:"id"`
	PaidAt            *string                       `json:"paid-at"`
	ServiceType       string                        `json:"service-type"`
	TargetName        *string                       `json:"target-name"`
	Description       *string                       `json:"description"`
	Amount            int                           `json:"amount"`
	AmountBreakdown   PaymentHistoryAmountBreakdown `json:"amount-breakdown"`
	PaymentMethod     *string                       `json:"payment-method"`
	Status            int                           `json:"status"`
	Refunded          bool                          `json:"refunded"`
	RefundedAt        *string                       `json:"refunded-at"`
	ContractStartDate *string                       `json:"contract-start-date"`
	ContractEndDate   *string                       `json:"contract-end-date"`
}

type InputGetPaymentHistory struct {
	ListOptions
	DateFrom    string
	DateTo      string
	AmountMin   *int
	AmountMax   *int
	ServiceType string
	DomainName  string
}

type OutputGetPaymentHistory struct {
	PaymentHistory []PaymentHistory
	Meta           PaginationMeta
}

func (c *Client) GetPaymentHistory(ctx context.Context, input *InputGetPaymentHistory) (*OutputGetPaymentHistory, error) {
	q := url.Values{}
	if input != nil {
		input.setTo(q)
		setString(q, "date-from", input.DateFrom)
		setString(q, "date-to", input.DateTo)
		setIntPtr(q, "amount-min", input.AmountMin)
		setIntPtr(q, "amount-max", input.AmountMax)
		setString(q, "service-type", input.ServiceType)
		setString(q, "domain-name", input.DomainName)
	}

	r, err := call[[]PaymentHistory](ctx, c, http.MethodGet, "/me/payment-history", q, nil)
	if err != nil {
		return nil, err
	}
	return &OutputGetPaymentHistory{PaymentHistory: r.Data, Meta: derefMeta(r.Meta)}, nil
}
