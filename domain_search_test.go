package muumuu

import (
	"context"
	"net/http"
	"testing"
)

func TestSearchDomains(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/domain-search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testQuery(t, r, "q=example&tlds=com%2Cjp")
		respond(t, w, http.StatusOK, `{"data":{"query":"example","results":[
			{"fqdn":"example.com","availability":"unavailable","unavailable-reason":"registered","domain-price":null,"service-fee":null,"tax":null,"total":null,"currency":"JPY"},
			{"fqdn":"example.jp","availability":"available","unavailable-reason":null,"domain-price":1000,"service-fee":100,"tax":110,"total":1210,"currency":"JPY"}
		],"truncated":false}}`)
	})

	got, err := c.SearchDomains(context.Background(), &InputSearchDomains{Q: "example", TLDs: []string{"com", "jp"}})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DomainSearchData{
		Query: "example",
		Results: []DomainSearchResult{
			{FQDN: "example.com", Availability: DomainAvailabilityUnavailable, UnavailableReason: String("registered"), Currency: "JPY"},
			{FQDN: "example.jp", Availability: DomainAvailabilityAvailable, DomainPrice: Int(1000), ServiceFee: Int(100), Tax: Int(110), Total: Int(1210), Currency: "JPY"},
		},
	})
}

func TestSearchDomains_RequiresQ(t *testing.T) {
	c, _ := setup(t)
	if _, err := c.SearchDomains(context.Background(), &InputSearchDomains{}); err == nil {
		t.Fatal("expected error")
	}
}
