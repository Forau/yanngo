package api_test

import (
	api "."
	"encoding/json"
	"testing"
)

func TestRequestEncoding(t *testing.T) {
	req, err := api.NewRequest("GET", "/test", &struct {
		Name   string
		Status float64 `json:"status"`
	}{Name: "TestName", Status: 42})
	if err != nil {
		t.Error(err)
	}
	t.Log(req)

	resMap := make(map[string]interface{})
	err = json.Unmarshal(req.Query, &resMap)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Res: %+v\n", resMap)
	if resMap["Name"] != "TestName" {
		t.Error("Expected 'TestName' as name, but got ", resMap["Name"])
	}
	if resMap["status"] != 42.0 {
		t.Errorf("Expected '42' as status, but got %+v as %T", resMap["status"], resMap["status"])
	}
}

func TestRequestResponse(t *testing.T) {
	perform := func(req *api.Request) (res *api.Response) {
		res = &api.Response{}
		if req.Action == "HEJ" {
			res.Success(&struct {
				Msg string
			}{"All is well"})
		} else {
			res.Fail(42, "We did not get HEJ")
		}
		return
	}

	req, err := api.NewRequest("HEJ", "/test", &struct {
		Data string
	}{Data: "TestData"})
	if err != nil {
		t.Error(err)
	}
	t.Log(req)

	res := perform(req)
	t.Log(res)

	if res.Error != nil {
		t.Error("Expected no error, but got ", res.Error)
	}

	req, err = api.NewRequest("WRONG", "/test", &struct{}{})
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
