package api

import (
	"encoding/json"
	"fmt"
)

type ErrorStatus int64

type ErrorHolder struct {
	Status  ErrorStatus
	Message string
}

func (eh *ErrorHolder) Error() string {
	return fmt.Sprintf("{\"status\": %d, \"message\": \"%s\"}", eh.Status, eh.Message)
}

type Response struct {
	Error   *ErrorHolder    // Error, if any. nil if not
	Payload json.RawMessage // Response payload as JSON
}

func (ar *Response) Fail(status ErrorStatus, msg string) {
	ar.Error = &ErrorHolder{Status: status, Message: msg}
}

func (ar *Response) Success(res interface{}) {
	ar.Error = nil // If we had an error, it is resolved now
	payload, err := json.Marshal(res)
	if err != nil {
		ar.Fail(-1, err.Error())
	} else {
		ar.Payload = payload
	}
}

func (ar *Response) IsError() bool {
	return ar.Error != nil
}

type Request struct {
	Action string          // GET or POST
	Path   string          // Request path
	Query  json.RawMessage // Will get unmashaled and converted to query or form params
}

func NewRequest(action, path string, query interface{}) (req *Request, err error) {
	req = &Request{Action: action, Path: path}
	if query != nil {
		req.Query, err = json.Marshal(query)
	}
	fmt.Printf("Request %+v\n", req)
	return
}

// Invoke the request. If there is an error, it should be set in the response struct
type Transport func(*Request) Response

type TransportHandler interface {
	Preform(*Request) Response
}

// Let the func implement the handler
func (p Transport) Preform(req *Request) Response {
	return p(req)
}
