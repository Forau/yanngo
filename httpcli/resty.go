// Copyright (c) 2016 Forau @ github.com. MIT License.

// Package httpcli contains http-based transports
package httpcli

import (
	"encoding/json"
	"github.com/Forau/yanngo/crypto"
	"github.com/Forau/yanngo/swagger"
	"gopkg.in/resty.v0" // https://github.com/go-resty/resty

	"os"

	"fmt"
	"time"
)

type RestError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (re RestError) Error() string {
	return fmt.Sprintf(`{"code": "%s", "message": "%s"}`, re.Code, re.Message)
}

func convertToStringMap(in map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k := range in {
		val := in[k]
		switch t := val.(type) {
		case string:
			res[k] = t
		case float64, float32:
			res[k] = fmt.Sprintf("%f", t)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			res[k] = fmt.Sprintf("%d", t)
		default:
			fmt.Printf("Dont know how to convert %T: %+v. Will try make it %v.\n", t, t, t)
			res[k] = fmt.Sprintf("%v", t)
		}
	}
	return res
}

// Same as in baseFeed.  TODO: Move to common
func makeDelayRetry(msg string) func() {
	resetTimer := time.Now().Add(time.Minute)
	counter := 0
	return func() {
		if time.Now().After(resetTimer) {
			counter = 0
		}
		resetTimer = time.Now().Add(time.Minute) // Reset in one minute, if no other call
		counter++
		var delay time.Duration
		switch counter {
		case 1:
			delay = 0
		case 2:
			delay = 5000
		default:
			delay = 30000
		}
		fmt.Printf(msg, delay)
		time.Sleep(delay * time.Millisecond)
	}
}

type RestClient struct {
	restyCli *resty.Client

	generate crypto.GenerateCredentials
	session  *swagger.Login

	lastSuccess time.Time

	retryDelayLogin func()

	waitLockTime time.Time
}

func NewRestClient(uri string, user, pass, pem []byte) *RestClient {
	var err error
	rc := &RestClient{retryDelayLogin: makeDelayRetry("Waiting %dms before trying to login again\n")}
	rc.generate, err = crypto.NewCredentialsGenerator(user, pass, pem)
	if err != nil {
		panic(err)
	}
	logFile, _ := os.OpenFile("go-resty.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	rc.restyCli = resty.New().SetLogger(logFile).SetDebug(true).SetHostURL(uri).SetHeaders(map[string]string{
		"Accept":          "application/json",
		"Accept-Language": "en",
		"User-Agent":      "YANNGO v0.0 (Yet Another NordNet GO - API)",
	})

	rc.restyCli.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		fmt.Printf(">>> %p >>> %s %s: %v\n", r, r.Method, r.URL, r.RawRequest.Body)
		return nil
	})
	rc.restyCli.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		fmt.Printf("<<< %p <<< %s [%s : %v]: %v\n", r.Request, r.Status(), r.Time(), r.Error(), r)
		if r.StatusCode() < 300 {
			rc.lastSuccess = time.Now()
		}
		return nil
	})

	go func() {
		// If we close, nil restCli
		for rc.restyCli != nil {
			time.Sleep(60000 * time.Millisecond)
			if time.Now().Unix() > (60 + rc.lastSuccess.Unix()) {
				rc.Execute("PUT", "login", nil) // Touch session
			}
		}
	}()

	return rc
}

func (rc *RestClient) Execute(method, path string, payload map[string]string) (json.RawMessage, error) {
	for time.Now().Before(rc.waitLockTime) {
		// We could count exact, but lets just loop every second and print in logs.
		fmt.Printf("Waiting nicely for nordnet. Now: %v, Allowed at: %v\n", time.Now().String(), rc.waitLockTime.String())
		time.Sleep(time.Second)
	}

	sess, err := rc.GetSession()
	if err != nil {
		fmt.Printf("Login error: %+v\n", err)
		return nil, err
	}

	// For very special calls.  Initially to get the session without calling the server again.
	if method == "SPECIAL" {
		if path == "session" {
			return json.Marshal(sess)
		}
		return nil, fmt.Errorf("No special command '%s'", path)
	}

	restError := RestError{}
	req := rc.restyCli.R().SetBasicAuth(sess.SessionKey, sess.SessionKey).SetError(&restError)
	if payload != nil {
		if method == "POST" || method == "PUT" {
			req.SetFormData(payload)
		} else {
			req.SetQueryParams(payload)
		}
	}

	fmt.Printf("About to execute %+v with %s on %s\n", req, method, path)
	resp, err := req.Execute(method, path)
	if err != nil {
		fmt.Println("HTTP-error: ", err)
		return nil, err
	}
	if restError.Code != "" {
		fmt.Println("REST-error: ", restError)
		if restError.Code == "NEXT_INVALID_SESSION" {
			if sess == rc.session {
				rc.session = nil
			}
			return rc.Execute(method, path, payload)
		} else if resp.StatusCode() == 429 { // Too Many Requests, please wait for 10 seconds before trying again
			rc.waitLockTime = time.Now().Add(time.Duration(10) * time.Second)
		}
		return nil, fmt.Errorf("%d: %s %s: %v", resp.StatusCode(), method, path, restError)
	}
	return resp.Body(), nil
}

func (rc *RestClient) GetSession() (*swagger.Login, error) {
	if sess := rc.session; sess != nil {
		return sess, nil // We store it in localy before the compare, so if it is reset concurrently, we wont return nil
	}
	// TODO: lock
	// defer UNLOCK
	rc.retryDelayLogin()

	auth, err := rc.generate()
	if err == nil {
		tmpSess := &swagger.Login{}
		_, err := rc.restyCli.R().
			SetFormData(map[string]string{"auth": auth, "service": "NEXTAPI"}).
			SetResult(&tmpSess).
			Post("/login")

		if err != nil {
			fmt.Println("Login error ", err)
			return nil, err
		}
		fmt.Printf("LOGIN %+v\n", tmpSess)
		rc.session = tmpSess
	}
	return rc.session, nil
}
