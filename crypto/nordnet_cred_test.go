package crypto_test

import (
	"github.com/Forau/yanngo/crypto"
	"testing"

	"io/ioutil"
	"os"

	"time"
)

func TestGenerateCredentials(t *testing.T) {
	file, err := os.Open("../NEXTAPI_TEST_public.pem")
	if err != nil {
		t.Fatal(err)
	}
	pemData, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("PEM: ", pemData)

	generate, err := crypto.NewCredentialsGenerator([]byte("user"), []byte("pass"), pemData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Generator: ", generate)

	cred, err := generate()

	if err != nil {
		t.Fatal(err)
	}
	t.Log("Cred: ", cred)
	if len(cred) < 32 {
		t.Error("Expected credential string to be longer, but was only ", len(cred), "b")
	}

	// Sleep one sec, to get a new timestamp
	time.Sleep(time.Second)

	cred2, err := generate()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Cred2: ", cred2)
	if len(cred2) < 32 {
		t.Error("Expected credential string to be longer, but was only ", len(cred2), "b")
	}

	if cred == cred2 {
		t.Error("Expected cred and cred2 to be different")
	}
}
