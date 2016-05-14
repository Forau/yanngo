package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"
)

// We will return a function, that gives us a freshly generated string each time
type GenerateCredentials func() (string, error)

func NewCredentialsGenerator(username, password, rawPem []byte) (gc GenerateCredentials, err error) {
	userBase64 := base64.StdEncoding.EncodeToString(username)
	passBase64 := base64.StdEncoding.EncodeToString(password)

	block, _ := pem.Decode(rawPem)
	if block == nil {
		err = fmt.Errorf("Could not Decode PEM")
		return
	}

	pubKeyVal, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return
	}

	rsaPubKey, ok := pubKeyVal.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("Could not make DER to an RSA PublicKey")
		return
	}

	// The function to generate new credentials
	gc = func() (string, error) {
		ms := time.Now().Unix() * 1000
		unixStr := strconv.FormatInt(ms, 10)
		timeBase64 := base64.StdEncoding.EncodeToString([]byte(unixStr))
		formated := fmt.Sprintf("%s:%s:%s", userBase64, passBase64, timeBase64)

		encr, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, []byte(formated))
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(encr), nil
	}
	return
}
