package api_test

import (
	"github.com/Forau/yanngo/api"
	"testing"
)

func TestGetAccounts(t *testing.T) {
	var transporth api.Transport = func(req *api.Request) (res api.Response) {
		if req.Command == api.AccountsCmd {
			res.Payload = []byte(`[{"accno": 12345}]`)
		}
		return
	}
	cli := api.NewApiClient(transporth)

	res, err := cli.Accounts()
	t.Log("Res: ", res, ", Err: ", err)
	if err != nil {
		t.Error(err)
	} else if len(res) < 1 {
		t.Error("Expected to get an account back")
	}
}
