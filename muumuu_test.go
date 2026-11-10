package muumuu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const testBasePath = "/api/v2"

func setup(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	api := http.NewServeMux()
	api.Handle(testBasePath+"/", http.StripPrefix(testBasePath, mux))
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)

	c, err := New("test-token", OptionBaseURL(server.URL+testBasePath), OptionHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	return c, mux
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("method = %s, want %s", r.Method, want)
	}
}

func testQuery(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.URL.RawQuery; got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
}

func testJSONBody(t *testing.T, r *http.Request, want string) {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	var got, exp any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("invalid request body %q: %v", b, err)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("body = %s, want %s", b, want)
	}
}

func respond(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := fmt.Fprint(w, body); err != nil {
		t.Fatal(err)
	}
}

func assertEqual(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func assertAPIError(t *testing.T, err error, status int, code string) {
	t.Helper()
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if e.StatusCode != status || e.Code != code {
		t.Errorf("error = (%d, %q), want (%d, %q)", e.StatusCode, e.Code, status, code)
	}
}

func mustTime(t *testing.T, s string) *time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return &v
}

func TestNew_DefaultsToProduction(t *testing.T) {
	c, err := New("token")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, c.baseURL.String(), BaseURLProduction)
}

func TestNew_RejectsNilHTTPClient(t *testing.T) {
	if _, err := New("token", OptionHTTPClient(nil)); err == nil {
		t.Fatal("expected error")
	}
}

func TestNew_RejectsInvalidBaseURL(t *testing.T) {
	if _, err := New("token", OptionBaseURL("://invalid")); err == nil {
		t.Fatal("expected error")
	}
}

func TestRequest_SendsAuthAndUserAgentHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/maintenances", func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Header.Get("Authorization"), "Bearer test-token")
		assertEqual(t, r.Header.Get("User-Agent"), "custom-ua")
		assertEqual(t, r.Header.Get("Accept"), "application/json")
		respond(t, w, http.StatusOK, `{"data":{"registrars":[],"registries":[],"services":[]}}`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	c, err := New("test-token", OptionBaseURL(server.URL+"/api/v2/"), OptionUserAgent("custom-ua"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetMaintenances(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRequest_ReturnsAPIErrorWithDetails(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusUnprocessableEntity, `{"error":{"code":"validation_error","message":"invalid","details":[{"field":"auto-renew-enabled","message":"credit card required"}]}}`)
	})

	_, err := c.UpdateDomain(context.Background(), "MU00000001", &InputUpdateDomain{AutoRenewEnabled: Bool(true)})
	assertAPIError(t, err, http.StatusUnprocessableEntity, ErrorCodeValidationError)

	var e *Error
	errors.As(err, &e)
	assertEqual(t, e.Details, []ErrorDetail{{Field: "auto-renew-enabled", Message: "credit card required"}})
	assertEqual(t, err.Error(), "muumuu-domain: 422 validation_error: invalid [auto-renew-enabled: credit card required]")
}

func TestRequest_ReturnsRateLimitErrorWithRetryAfter(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/maintenances", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		respond(t, w, http.StatusTooManyRequests, `{"error":{"code":"rate_limit_exceeded","message":"Rate limit exceeded"}}`)
	})

	_, err := c.GetMaintenances(context.Background())
	if !IsRateLimited(err) {
		t.Fatalf("IsRateLimited(%v) = false", err)
	}
	var e *Error
	errors.As(err, &e)
	assertEqual(t, e.RetryAfter, 60*time.Second)
}

func TestRequest_ReturnsErrorForNonJSONBody(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/maintenances", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	_, err := c.GetMaintenances(context.Background())
	assertAPIError(t, err, http.StatusBadGateway, "")
	assertEqual(t, err.Error(), "muumuu-domain: 502: Bad Gateway")
}

func TestRequest_ReturnsErrorForMalformedSuccessBody(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/maintenances", func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, `{"data":`)
	})

	if _, err := c.GetMaintenances(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestRequest_ReturnsErrorForEmptySuccessBody(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/MU00000001", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if _, err := c.GetDomain(context.Background(), "MU00000001"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRequest_EscapesPathSegmentOnce(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/domains/a%2Fb", func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.EscapedPath(), "/me/domains/a%2Fb")
		respond(t, w, http.StatusOK, `{"data":{"id":"a/b"}}`)
	})

	got, err := c.GetDomain(context.Background(), "a/b")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got.ID, "a/b")
}

func TestRequest_HonorsContextCancellation(t *testing.T) {
	c, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.GetMaintenances(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestPaginationMeta_HasNext(t *testing.T) {
	cases := []struct {
		meta PaginationMeta
		want bool
	}{
		{PaginationMeta{Total: 21, Page: 1, PageSize: 20}, true},
		{PaginationMeta{Total: 20, Page: 1, PageSize: 20}, false},
		{PaginationMeta{Total: 0, Page: 1, PageSize: 20}, false},
		{PaginationMeta{}, false},
	}
	for _, tc := range cases {
		assertEqual(t, tc.meta.HasNext(), tc.want)
	}
}
