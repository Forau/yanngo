package api_test

import (
	"github.com/Forau/yanngo/api"
	"testing"
)

func TestRequestEncoding(t *testing.T) {
	req, err := api.NewRequest("TestRequest", map[string]string{"Name": "TestName", "status": "42"})
	if err != nil {
		t.Error(err)
	}
	t.Log(req)

	resMap := req.Params
	t.Logf("Res: %+v\n", resMap)
	if resMap["Name"] != "TestName" {
		t.Error("Expected 'TestName' as name, but got ", resMap["Name"])
	}
	if resMap["status"] != "42" {
		t.Errorf("Expected '42' as status, but got %+v as %T", resMap["status"], resMap["status"])
	}
}

func TestRequestResponse(t *testing.T) {
	perform := func(req *api.Request) (res api.Response) {
		if req.Command == "TestRequest" {
			res.Success(&struct {
				Msg string
			}{"All is well"})
		} else {
			res.Fail(42, "We did not get HEJ")
		}
		return
	}

	req, err := api.NewRequest("TestRequest", map[string]string{"Data": "TestData"})
	if err != nil {
		t.Error(err)
	}
	t.Log(req)

	res := perform(req)
	t.Log(res)

	if res.Error != nil {
		t.Error("Expected no error, but got ", res.Error)
	}

	req, err = api.NewRequest("WRONGRequest", nil)
	if err != nil {
		t.Error(err)
	}
	t.Log(req)

	res = perform(req)
	t.Log(res)

	if res.Error == nil {
		t.Error("Expected error, but got ", res.Error)
	} else if res.Error.Status != 42 {
		t.Error("Expected status to be 42, but was ", res.Error.Status)
	}

}
