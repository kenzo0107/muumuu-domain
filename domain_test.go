package muumuu

import (
	"context"
	"net/http"
	"testing"
)

const domainJSON = `{
	"id": "MU00000001",
	"sld": "example",
	"tld": "com",
	"fqdn": "example.com",
	"state": "active",
	"setup-state": "acquired",
	"registrar": "onamae",
	"whois-proxy-enabled": true,
	"auto-renew-enabled": false,
	"is-japanese-domain": false,
	"contract": {
		"id": "MC00000001",
		"state": "new",
		"term": 1,
		"start-date": "2024-01-01T00:00:00+09:00",
		"end-date": "2025-01-01T00:00:00+09:00",
		"payment-method": "credit",
		"auto-renew-enabled": false
	}
}`

func wantDomain(t *testing.T) *Domain {
	return &Domain{
		ID:                "MU00000001",
		SLD:               "example",
		TLD:               "com",
		FQDN:              "example.com",
		State:             "active",
		SetupState:        "acquired",
		Registrar:         "onamae",
		WhoisProxyEnabled: Bool(true),
		AutoRenewEnabled:  Bool(false),
		IsJapaneseDomain:  Bool(false),
		Contract: &DomainContract{
			ID:               "MC00000001",
			State:            "new",
			Term:             Int(1),
			StartDate:        mustTime(t, "2024-01-01T00:00:00+09:00"),
			EndDate:          mustTime(t, "2025-01-01T00:00:00+09:00"),
			PaymentMethod:    String("credit"),
			AutoRenewEnabled: Bool(false),
		},
	}
}

func TestGetDomains(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testQuery(t, r, "fqdn=example.com&page=2&page-size=50&q=exa&sort=status-group&state=active&status-group=contracted%2Cexpired-recoverable")
		respond(t, w, http.StatusOK, `{"data":[`+domainJSON+`],"meta":{"total":51,"page":2,"page-size":50}}`)
	})

	got, err := c.GetDomains(context.Background(), &InputGetDomains{
		ListOptions: ListOptions{Page: 2, PageSize: 50},
		FQDN:        "example.com",
		State:       "active",
		StatusGroup: []string{"contracted", "expired-recoverable"},
		Sort:        "status-group",
		Q:           "exa",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &OutputGetDomains{
		Domains: []Domain{*wantDomain(t)},
		Meta:    PaginationMeta{Total: 51, Page: 2, PageSize: 50},
	})
}

func TestGetDomains_WithoutInputSendsNoQuery(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains", func(w http.ResponseWriter, r *http.Request) {
		testQuery(t, r, "")
		respond(t, w, http.StatusOK, `{"data":[],"meta":{"total":0,"page":1,"page-size":20}}`)
	})

	got, err := c.GetDomains(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, len(got.Domains), 0)
}

func TestGetDomain(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		respond(t, w, http.StatusOK, `{"data":`+domainJSON+`}`)
	})

	got, err := c.GetDomain(context.Background(), "MU00000001")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, wantDomain(t))
}

func TestGetDomain_NotFound(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU99999999", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusNotFound, `{"error":{"code":"not_found","message":"Domain not found"}}`)
	})

	_, err := c.GetDomain(context.Background(), "MU99999999")
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
	assertAPIError(t, err, http.StatusNotFound, ErrorCodeNotFound)
}

func TestUpdateDomain(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		testJSONBody(t, r, `{"auto-renew-enabled":false}`)
		respond(t, w, http.StatusOK, `{"data":`+domainJSON+`}`)
	})

	got, err := c.UpdateDomain(context.Background(), "MU00000001", &InputUpdateDomain{AutoRenewEnabled: Bool(false)})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, wantDomain(t))
}
