//go:build live

package muumuu

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func liveClient(t *testing.T) *Client {
	t.Helper()
	token := os.Getenv("MUUMUU_DOMAIN_TOKEN")
	if token == "" {
		t.Fatal("MUUMUU_DOMAIN_TOKEN is required")
	}
	baseURL := os.Getenv("MUUMUU_DOMAIN_BASE_URL")
	if baseURL == "" {
		baseURL = BaseURLSandbox
	}
	c, err := New(token, OptionBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLiveGetMaintenances(t *testing.T) {
	if _, err := liveClient(t).GetMaintenances(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestLiveGetDomains(t *testing.T) {
	out, err := liveClient(t).GetDomains(context.Background(), &InputGetDomains{ListOptions: ListOptions{Page: 1, PageSize: 100}})
	if err != nil {
		t.Fatal(err)
	}
	if out.Meta.Page != 1 || out.Meta.PageSize != 100 {
		t.Errorf("meta = %+v", out.Meta)
	}
}

func TestLiveGetDomain_NotFound(t *testing.T) {
	_, err := liveClient(t).GetDomain(context.Background(), "MU99999999")
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
}

func TestLiveSearchDomains(t *testing.T) {
	out, err := liveClient(t).SearchDomains(context.Background(), &InputSearchDomains{Q: "example", TLDs: []string{"com", "jp"}})
	if err != nil {
		t.Fatal(err)
	}
	if out.Query != "example" || len(out.Results) == 0 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestLiveGetDNSRecordsByDomainFQDN_UnownedDomain(t *testing.T) {
	_, err := liveClient(t).GetDNSRecordsByDomainFQDN(context.Background(), &InputGetDNSRecordsByDomainFQDN{DomainFQDN: "example.com"})
	assertAPIError(t, err, http.StatusNotFound, ErrorCodeNotFound)
}

func TestLiveDNSRecordLifecycle(t *testing.T) {
	domainID := os.Getenv("MUUMUU_DOMAIN_TEST_DOMAIN_ID")
	if domainID == "" {
		t.Skip("MUUMUU_DOMAIN_TEST_DOMAIN_ID is not set")
	}
	c := liveClient(t)
	ctx := context.Background()

	domain, err := c.GetDomain(ctx, domainID)
	if err != nil {
		t.Fatal(err)
	}

	created, err := c.CreateDNSRecord(ctx, domainID, &InputCreateDNSRecord{
		FQDN:  "muumuu-go-live-test." + domain.FQDN,
		Type:  DNSRecordTypeTXT,
		Value: "created",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.DeleteDNSRecord(ctx, domainID, created.ID); err != nil && !IsNotFound(err) {
			t.Errorf("cleanup: %v", err)
		}
	})

	updated, err := c.UpdateDNSRecord(ctx, domainID, created.ID, &InputUpdateDNSRecord{Value: String("updated")})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, updated.Value, "updated")

	got, err := c.GetDNSRecord(ctx, domainID, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got.Value, "updated")

	if err := c.DeleteDNSRecord(ctx, domainID, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetDNSRecord(ctx, domainID, created.ID); !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
}
