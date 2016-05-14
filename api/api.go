package api

import (
	"encoding/json"
	"github.com/Forau/yanngo/swagger"
)

type ApiClient struct {
	ph TransportHandler
}

func NewApiClient(ph Transport) *ApiClient {
	return &ApiClient{ph: ph}
}

func (ac *ApiClient) LoginStatus() (res swagger.LoggedInStatus, err error) {
	req, err := NewRequest("GET", "/login", nil)
	if err == nil {
		resp := ac.ph.Preform(req)
		if resp.Error != nil {
			err = resp.Error
		} else {
			err = json.Unmarshal(resp.Payload, &res)
		}
	}
	return
}
