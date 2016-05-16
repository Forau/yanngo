package api

import (
	"encoding/json"
	"github.com/Forau/yanngo/swagger"

	//  "fmt"
)

type ApiClient struct {
	ph TransportHandler
}

func NewApiClient(ph Transport) *ApiClient {
	return &ApiClient{ph: ph}
}

func (ac *ApiClient) handleRequest(command RequestCommand, query interface{}, res interface{}) error {
	req, err := NewRequest(command, query)
	if err != nil {
		return err
	}
	resp := ac.ph.Preform(req)
	if resp.Error != nil {
		return resp.Error
	} else {
		return json.Unmarshal(resp.Payload, res)
	}
}

func (ac *ApiClient) Accounts() (res []swagger.Account, err error) {
	err = ac.handleRequest(AccountsCmd, nil, &res)
	return
}

type accountQuery struct {
	Accno int64
}

func (ac *ApiClient) Account(accno int64) (res swagger.AccountInfo, err error) {
	err = ac.handleRequest(AccountCmd, &accountQuery{accno}, &res)
	return
}
