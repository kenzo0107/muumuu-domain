package muumuu

import (
	"context"
	"net/http"
	"testing"
)

func TestGetPersonalAccessTokens(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/me/personal-access-tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		respond(t, w, http.StatusOK, `{"data":[{"id":1,"name":"terraform","token-prefix":"muu_pat_abcdefgh…","scopes":["dns:read","dns:write"],"expires-at":null,"last-used-at":"2024-01-01T00:00:00Z","status":"active"}]}`)
	})

	got, err := c.GetPersonalAccessTokens(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, []PersonalAccessToken{{
		ID:          1,
		Name:        "terraform",
		TokenPrefix: "muu_pat_abcdefgh…",
		Scopes:      []string{"dns:read", "dns:write"},
		LastUsedAt:  mustTime(t, "2024-01-01T00:00:00Z"),
		Status:      "active",
	}})
}

func TestDeletePersonalAccessToken(t *testing.T) {
	c, mux := setup(t)
	called := false
	mux.HandleFunc("/me/personal-access-tokens/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.DeletePersonalAccessToken(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, called, true)
}
