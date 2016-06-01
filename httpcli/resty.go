package httpcli

import (
	"encoding/json"
	"github.com/Forau/yanngo/crypto"
	"github.com/Forau/yanngo/swagger"
	"gopkg.in/resty.v0" // https://github.com/go-resty/resty

	"os"

	"fmt"
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

type RestClient struct {
	restyCli *resty.Client

	generate crypto.GenerateCredentials
	session  *swagger.Login
}

func NewRestClient(uri string, user, pass, pem []byte) *RestClient {
	var err error
	rc := &RestClient{}
	rc.generate, err = crypto.NewCredentialsGenerator(user, pass, pem)
	if err != nil {
		panic(err)
	}
	logFile, _ := os.OpenFile("go-resty.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	rc.restyCli = resty.New().SetLogger(logFile).SetDebug(true).SetHostURL(uri).SetHeaders(map[string]string{
		"Accept":          "application/json",
		"Accept-Language": "en",
	}).SetError(RestError{})

	rc.restyCli.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		fmt.Printf(">>> %p >>> %s %s: %v\n", r, r.Method, r.URL, r.RawRequest.Body)
		return nil
	})
	rc.restyCli.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		fmt.Printf("<<< %p <<< %s [%s]: %v\n", r.Request, r.Status(), r.Time(), r)
		return nil
	})

	return rc
}

func (rc *RestClient) Execute(method, path string, payload map[string]interface{}) (json.RawMessage, error) {
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

	req := rc.restyCli.R().SetBasicAuth(sess.SessionKey, sess.SessionKey)
	if payload != nil {
		if method == "POST" || method == "PUT" {
			req.SetFormData(convertToStringMap(payload))
		} else {
			req.SetQueryParams(convertToStringMap(payload))
		}
	}
	fmt.Printf("About to execute %+v with %s on %s\n", req, method, path)
	resp, err := req.Execute(method, path)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (rc *RestClient) GetSession() (*swagger.Login, error) {
	if sess := rc.session; sess != nil {
		return sess, nil // We store it in localy before the compare, so if it is reset concurrently, we wont return nil
	}
	// TODO: lock
	// defer UNLOCK

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
