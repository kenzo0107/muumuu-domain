package muumuu

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestGetDNSRecords(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testQuery(t, r, "fqdn=www.example.com&page=1&page-size=100&type=MX")
		respond(t, w, http.StatusOK, `{"data":[{
			"id": 12345,
			"fqdn": "www.example.com.",
			"type": "MX",
			"value": "mail.example.com.",
			"ttl": 3600,
			"priority": 10,
			"created-at": "2024-01-01T00:00:00Z",
			"updated-at": "2024-01-02T00:00:00Z"
		}],"meta":{"total":1,"page":1,"page-size":100}}`)
	})

	got, err := c.GetDNSRecords(context.Background(), "MU00000001", &InputGetDNSRecords{
		ListOptions: ListOptions{Page: 1, PageSize: 100},
		Type:        DNSRecordTypeMX,
		FQDN:        "www.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &OutputGetDNSRecords{
		DNSRecords: []DNSRecord{{
			ID:        12345,
			FQDN:      "www.example.com.",
			Type:      "MX",
			Value:     "mail.example.com.",
			TTL:       3600,
			Priority:  Int(10),
			CreatedAt: mustTime(t, "2024-01-01T00:00:00Z"),
			UpdatedAt: mustTime(t, "2024-01-02T00:00:00Z"),
		}},
		Meta: PaginationMeta{Total: 1, Page: 1, PageSize: 100},
	})
}

func TestGetDNSRecordsByDomainFQDN(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/dns-records", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testQuery(t, r, "domain-fqdn=example.com&type=A")
		respond(t, w, http.StatusOK, `{"data":[{"id":1,"fqdn":"example.com.","type":"A","value":"192.0.2.1","ttl":3600}],"meta":{"total":1,"page":1,"page-size":20}}`)
	})

	got, err := c.GetDNSRecordsByDomainFQDN(context.Background(), &InputGetDNSRecordsByDomainFQDN{
		InputGetDNSRecords: InputGetDNSRecords{Type: DNSRecordTypeA},
		DomainFQDN:         "example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got.DNSRecords, []DNSRecord{{ID: 1, FQDN: "example.com.", Type: "A", Value: "192.0.2.1", TTL: 3600}})
}

func TestGetDNSRecordsByDomainFQDN_RequiresDomainFQDN(t *testing.T) {
	c, _ := setup(t)
	if _, err := c.GetDNSRecordsByDomainFQDN(context.Background(), &InputGetDNSRecordsByDomainFQDN{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetDNSRecord_FindsRecordOnLaterPage(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		assertEqual(t, r.URL.Query().Get("page-size"), "100")
		id := map[string]int{"1": 1, "2": 2}[page]
		respond(t, w, http.StatusOK, fmt.Sprintf(`{"data":[{"id":%d,"fqdn":"example.com.","type":"A","value":"192.0.2.%d","ttl":3600}],"meta":{"total":101,"page":%s,"page-size":100}}`, id, id, page))
	})

	got, err := c.GetDNSRecord(context.Background(), "MU00000001", 2)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DNSRecord{ID: 2, FQDN: "example.com.", Type: "A", Value: "192.0.2.2", TTL: 3600})
}

func TestGetDNSRecord_ReturnsNotFoundWhenAbsent(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, `{"data":[{"id":1,"fqdn":"example.com.","type":"A","value":"192.0.2.1","ttl":3600}],"meta":{"total":1,"page":1,"page-size":100}}`)
	})

	_, err := c.GetDNSRecord(context.Background(), "MU00000001", 999)
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
}

func TestCreateDNSRecord(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{"fqdn":"example.com","type":"MX","value":"mail.example.com.","priority":10}`)
		respond(t, w, http.StatusCreated, `{"data":{"id":12345,"fqdn":"example.com.","type":"MX","value":"mail.example.com.","ttl":3600,"priority":10}}`)
	})

	got, err := c.CreateDNSRecord(context.Background(), "MU00000001", &InputCreateDNSRecord{
		FQDN:     "example.com",
		Type:     DNSRecordTypeMX,
		Value:    "mail.example.com.",
		Priority: Int(10),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &DNSRecord{ID: 12345, FQDN: "example.com.", Type: "MX", Value: "mail.example.com.", TTL: 3600, Priority: Int(10)})
}

func TestCreateDNSRecord_Conflict(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusConflict, `{"error":{"code":"conflict","message":"Record already exists"}}`)
	})

	_, err := c.CreateDNSRecord(context.Background(), "MU00000001", &InputCreateDNSRecord{FQDN: "example.com", Type: "A", Value: "192.0.2.1"})
	assertAPIError(t, err, http.StatusConflict, ErrorCodeConflict)
}

func TestUpdateDNSRecord(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		testJSONBody(t, r, `{"value":"192.0.2.2"}`)
		respond(t, w, http.StatusOK, `{"data":{"id":12345,"fqdn":"example.com.","type":"A","value":"192.0.2.2","ttl":3600}}`)
	})

	got, err := c.UpdateDNSRecord(context.Background(), "MU00000001", 12345, &InputUpdateDNSRecord{Value: String("192.0.2.2")})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got.Value, "192.0.2.2")
}

func TestDeleteDNSRecord(t *testing.T) {
	c, mux := setup(t)
	called := false
	mux.HandleFunc("/me/domains/MU00000001/dns-records/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.DeleteDNSRecord(context.Background(), "MU00000001", 12345); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, called, true)
}

func TestDeleteDNSRecord_NotFound(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001/dns-records/12345", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusNotFound, `{"error":{"code":"not_found","message":"Record not found"}}`)
	})

	if err := c.DeleteDNSRecord(context.Background(), "MU00000001", 12345); !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
}
