package httpcli

import (
	"encoding/json"
	"github.com/Forau/yanngo/crypto"
	"github.com/Forau/yanngo/swagger"
	"gopkg.in/resty.v0" // https://github.com/go-resty/resty

	"os"

	"fmt"

	"net/http/httputil"
)

type RestError struct {
	Msg string
}

func (re RestError) Error() string {
	return re.Msg
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
	})

	logDataOrError := func(data []byte, err error) {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(data))
		}
	}

	rc.restyCli.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		logDataOrError(httputil.DumpRequest(r.RawRequest, true))
		return nil
	})
	rc.restyCli.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		logDataOrError(httputil.DumpResponse(r.RawResponse, true))
		return nil
	})

	return rc
}

func (rc *RestClient) Execute(method, path string, payload map[string]interface{}) (json.RawMessage, error) {
	// Allow for one retry
	for i := 1; i > 0; i-- {
		sess, err := rc.GetSession()
		if err != nil {
			fmt.Printf("Login error: %+v\n", err)
			return nil, err
		}
		req := rc.restyCli.R()
		req.SetBasicAuth(sess.SessionKey, sess.SessionKey)

		if payload != nil {
			req.SetFormData(convertToStringMap(payload))
		}
		fmt.Printf("About to execute %+v with %s on %s\n", req, method, path)
		resp, err := req.Execute(method, path)
		if err != nil {
			fmt.Printf("Got error %+v\n", err)
			return nil, err
		}
		fmt.Printf("Got response %+v\n", resp)

		switch resp.StatusCode() {
		case 401:
			rc.session = nil // To force a login
		case 200, 201, 202, 203, 204:
			return resp.Body(), nil
		default:
			// All errors not specifically taken care of before.
			err := &RestError{string(resp.Body())}
			return nil, err
		}
	}
	return nil, &RestError{"Too many resends"}
}

func (rc *RestClient) GetSession() (*swagger.Login, error) {
	if sess := rc.session; sess != nil {
		return sess, nil // We store it in localy before the compare, so if it is reset concurrently, we wont return nil
	}
	// TODO: lock
	// defer UNLOCK

	auth, err := rc.generate()
	if err == nil {
		resp, err := rc.restyCli.R().
			SetFormData(map[string]string{"auth": auth, "service": "NEXTAPI"}).
			Post("/login")

		if err != nil {
			fmt.Println("Login error ", err)
			return nil, err
		}
		fmt.Printf("LOGIN %+v\n", resp)
		tmpSess := &swagger.Login{}
		err = json.Unmarshal(resp.Body(), tmpSess)
		if err != nil {
			return nil, err
		}
		rc.session = tmpSess
	}

	return rc.session, nil
}
