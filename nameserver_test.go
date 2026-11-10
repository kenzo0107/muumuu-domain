package muumuu

import (
	"context"
	"net/http"
	"testing"
)

func TestGetNameservers(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/nameservers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		respond(t, w, http.StatusOK, `{"data":{"domain-id":"MU00000001","setup-type":"muumuu_dns","nameservers":["ns01.muumuu-domain.com","ns02.muumuu-domain.com"]}}`)
	})

	got, err := c.GetNameservers(context.Background(), "MU00000001")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &NameserverSettings{
		DomainID:    "MU00000001",
		SetupType:   NameserverSetupTypeMuumuuDNS,
		Nameservers: []string{"ns01.muumuu-domain.com", "ns02.muumuu-domain.com"},
	})
}

func TestUpdateNameservers(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/nameservers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		testJSONBody(t, r, `{"setup-type":"custom","nameservers":["ns1.example.net","ns2.example.net"]}`)
		respond(t, w, http.StatusOK, `{"data":{"domain-id":"MU00000001","setup-type":"custom","nameservers":["ns1.example.net","ns2.example.net"]}}`)
	})

	got, err := c.UpdateNameservers(context.Background(), "MU00000001", &InputUpdateNameservers{
		SetupType:   NameserverSetupTypeCustom,
		Nameservers: []string{"ns1.example.net", "ns2.example.net"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got.Nameservers, []string{"ns1.example.net", "ns2.example.net"})
}

func TestUpdateNameservers_Conflict(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/nameservers", func(w http.ResponseWriter, r *http.Request) {
		testJSONBody(t, r, `{"setup-type":"muumuu_dns"}`)
		respond(t, w, http.StatusConflict, `{"error":{"code":"existing_dns_records","message":"existing records"}}`)
	})

	_, err := c.UpdateNameservers(context.Background(), "MU00000001", &InputUpdateNameservers{SetupType: NameserverSetupTypeMuumuuDNS})
	assertAPIError(t, err, http.StatusConflict, "existing_dns_records")
}
