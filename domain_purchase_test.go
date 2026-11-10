package muumuu

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestQuoteDomainPurchase(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domain-purchase/quote", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{"fqdn":"example.com","term":2}`)
		respond(t, w, http.StatusOK, `{"data":{
			"fqdn":"example.com","availability":"available","term":2,
			"domain-price":2000,"service-fee":200,"tax":220,"total":2420,"currency":"JPY",
			"credit-card-registered":true,"credit-card-registration-url":null,
			"purchase-token":{"token":"tok","expires-at":"2024-01-01T00:10:00Z"}
		}}`)
	})

	got, err := c.QuoteDomainPurchase(context.Background(), &InputQuoteDomainPurchase{FQDN: "example.com", Term: Int(2)})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DomainPurchaseQuote{
		FQDN:                 "example.com",
		Availability:         DomainAvailabilityAvailable,
		Term:                 2,
		DomainPrice:          Int(2000),
		ServiceFee:           Int(200),
		Tax:                  Int(220),
		Total:                Int(2420),
		Currency:             "JPY",
		CreditCardRegistered: true,
		PurchaseToken:        &PurchaseToken{Token: "tok", ExpiresAt: *mustTime(t, "2024-01-01T00:10:00Z")},
	})
}

func TestPurchaseDomain(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domain-purchase", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{"purchase-token":"tok"}`)
		respond(t, w, http.StatusCreated, `{"data":{"purchase-id":42,"checkout-url":"https://muumuu-domain.com/checkout/mcp/abc","status":"pending","fqdn":"example.com","term":1,"expires-at":"2024-01-01T00:10:00Z"}}`)
	})

	got, err := c.PurchaseDomain(context.Background(), &InputPurchaseDomain{PurchaseToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DomainPurchase{
		PurchaseID:  42,
		CheckoutURL: "https://muumuu-domain.com/checkout/mcp/abc",
		Status:      PurchaseStatusPending,
		FQDN:        "example.com",
		Term:        1,
		ExpiresAt:   *mustTime(t, "2024-01-01T00:10:00Z"),
	})
}

func TestPurchaseDomainDirect(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domain-purchase/direct", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{"purchase-token":"tok"}`)
		respond(t, w, http.StatusCreated, `{"data":{"purchase-id":"TP00000001","status":"completed","fqdn":"example.com","term":1,"domain-id":"MU00000001","contract-id":"MC00000001","completed-at":"2024-01-01"}}`)
	})

	got, err := c.PurchaseDomainDirect(context.Background(), &InputPurchaseDomain{PurchaseToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DirectDomainPurchase{
		PurchaseID:  "TP00000001",
		Status:      PurchaseStatusCompleted,
		FQDN:        "example.com",
		Term:        1,
		DomainID:    "MU00000001",
		ContractID:  "MC00000001",
		CompletedAt: "2024-01-01",
	})
}

func TestPurchaseDomainDirect_RegistrationFailedIncludesPurchaseID(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domain-purchase/direct", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusBadGateway, `{"error":{"code":"domain_registration_failed","message":"failed","purchase-id":"TP00000001","fqdn":"example.com"}}`)
	})

	_, err := c.PurchaseDomainDirect(context.Background(), &InputPurchaseDomain{PurchaseToken: "tok"})
	assertAPIError(t, err, http.StatusBadGateway, "domain_registration_failed")

	var e *Error
	errors.As(err, &e)
	assertEqual(t, e.PurchaseID, "TP00000001")
	assertEqual(t, e.FQDN, "example.com")
}

func TestGetDomainPurchaseStatus(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domain-purchase/42/status", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		respond(t, w, http.StatusOK, `{"data":{"purchase-id":42,"status":"failed","fqdn":"example.com","term":1,"domains":[
			{"domain-name":"example.com","term":1,"status":"failed","domain-id":null,"contract-id":null,"failure-reason":"payment_failed"}
		]}}`)
	})

	got, err := c.GetDomainPurchaseStatus(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DomainPurchaseStatus{
		PurchaseID: 42,
		Status:     PurchaseStatusFailed,
		FQDN:       "example.com",
		Term:       1,
		Domains: []DomainPurchaseStatusDomain{
			{DomainName: "example.com", Term: 1, Status: PurchaseStatusFailed, FailureReason: String("payment_failed")},
		},
	})
}
