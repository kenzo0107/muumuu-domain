package muumuu

import (
	"context"
	"net/http"
	"time"
)

type RegistrarMaintenance struct {
	Name          string    `json:"name"`
	BeginAt       time.Time `json:"begin-at"`
	EndAt         time.Time `json:"end-at"`
	InformationID *int      `json:"information-id"`
}

type RegistryMaintenance struct {
	TLDs          []string  `json:"tlds"`
	BeginAt       time.Time `json:"begin-at"`
	EndAt         time.Time `json:"end-at"`
	InformationID *int      `json:"information-id"`
}

type ServiceMaintenance struct {
	Name          string    `json:"name"`
	BeginAt       time.Time `json:"begin-at"`
	EndAt         time.Time `json:"end-at"`
	InformationID *int      `json:"information-id"`
}

type Maintenances struct {
	Registrars []RegistrarMaintenance `json:"registrars"`
	Registries []RegistryMaintenance  `json:"registries"`
	Services   []ServiceMaintenance   `json:"services"`
}

func (c *Client) GetMaintenances(ctx context.Context) (*Maintenances, error) {
	r, err := call[Maintenances](ctx, c, http.MethodGet, "/maintenances", nil, nil)
	if err != nil {
		return nil, err
	}
	return &r.Data, nil
}
