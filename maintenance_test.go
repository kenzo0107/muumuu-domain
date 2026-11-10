package muumuu

import (
	"context"
	"net/http"
	"testing"
)

func TestGetMaintenances(t *testing.T) {
	c, mux := setup(t)
	mux.HandleFunc("/maintenances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		respond(t, w, http.StatusOK, `{"data":{
			"registrars":[{"name":"jprs","begin-at":"2024-01-01T00:00:00+09:00","end-at":"2024-01-01T06:00:00+09:00","information-id":123}],
			"registries":[{"tlds":["com","net"],"begin-at":"2024-01-02T00:00:00+09:00","end-at":"2024-01-02T06:00:00+09:00","information-id":null}],
			"services":[]
		}}`)
	})

	got, err := c.GetMaintenances(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, got, &Maintenances{
		Registrars: []RegistrarMaintenance{{
			Name:          "jprs",
			BeginAt:       *mustTime(t, "2024-01-01T00:00:00+09:00"),
			EndAt:         *mustTime(t, "2024-01-01T06:00:00+09:00"),
			InformationID: Int(123),
		}},
		Registries: []RegistryMaintenance{{
			TLDs:    []string{"com", "net"},
			BeginAt: *mustTime(t, "2024-01-02T00:00:00+09:00"),
			EndAt:   *mustTime(t, "2024-01-02T06:00:00+09:00"),
		}},
		Services: []ServiceMaintenance{},
	})
}
