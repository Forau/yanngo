package api_test

import (
	"github.com/Forau/yanngo/api"
	"testing"
)

func TestLoggedInStatus(t *testing.T) {
	cli := api.NewApiClient(func(req *api.Request) (res api.Response) {
		if req.Path == "/login" {
			res.Payload = []byte(`{"logged_in": true}`)
		}
		return
	})

	res, err := cli.LoginStatus()
	t.Log("Res: ", res, ", Err: ", err)
	if err != nil {
		t.Error(err)
	} else if !res.LoggedIn {
		t.Error("Expected LoginStatus to be true")
	}
}
