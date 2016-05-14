package api_test

import (
	"github.com/Forau/yanngo/api"
	"testing"
)

func TestGetAccounts(t *testing.T) {
	cli := api.NewApiClient(func(req *api.Request) (res api.Response) {
		if req.Command == "Accounts" {
			res.Payload = []byte(`[{"accno": 12345}]`)
		}
		return
	})

	res, err := cli.Accounts()
	t.Log("Res: ", res, ", Err: ", err)
	if err != nil {
		t.Error(err)
	} else if len(res) < 1 {
		t.Error("Expected to get an account back")
	}
}
