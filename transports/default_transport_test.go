package transports_test

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/transports"
	"testing"

	"io/ioutil"
	"os"
)

var pemData []byte

func init() {
	file, err := os.Open("../NEXTAPI_TEST_public.pem")
	if err != nil {
		panic(err)
	}
	pemData, err = ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
}

func TestGetAccounts(t *testing.T) {
	tr, err := transports.NewDefaultTransport("end", []byte("kalle"), []byte("hemlig"), pemData)
	if err != nil {
		t.Fatal(err)
	}
	req, err := api.NewRequest(api.DeleteOrderCmd, map[string]int{"Accno": 123, "Id": 321})
	if err != nil {
		t.Fatal(err)
	}

	res := tr(req)
	t.Log(res)
}
