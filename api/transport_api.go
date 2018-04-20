package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ErrorStatus int64

type ErrorHolder struct {
	Status  ErrorStatus `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
}

func (eh *ErrorHolder) Error() string {
	return fmt.Sprintf("{\"status\": %d, \"message\": \"%s\"}", eh.Status, eh.Message)
}

type Response struct {
	Error   *ErrorHolder    `json:"error,omitempty"`   // Error, if any. nil if not
	Payload json.RawMessage `json:"payload,omitempty"` // Response payload as JSON
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

func (ar *Response) Unmarshal(ob interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(ar.Payload))
	dec.UseNumber()
	return dec.Decode(ob)
	//	return json.Unmarshal(ar.Payload, ob)
}

func (ar *Response) Marshal(ob interface{}) (err error) {
	ar.Payload, err = json.Marshal(ob)
	return
}

func (ar *Response) String() string {
	//return fmt.Sprintf("Response{error: %+v, payload: %s}", ar.Error, string(ar.Payload))
	b, err := json.Marshal(ar)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(b)
}

type RequestCommand string

// The commands that we support.  Will be mapped in the low level TransportHandler
const (
	// Internally used.  A well behaving transport should return the commands it can handle
	TransportRespondsToCmd RequestCommand = "TransportRespondsTo"

	SessionCmd                     RequestCommand = "Session"
	AccountsCmd                    RequestCommand = "Accounts"
	AccountCmd                     RequestCommand = "Account"
	AccountLedgersCmd              RequestCommand = "AccountLedgers"
	AccountOrdersCmd               RequestCommand = "AccountOrders"
	CreateOrderCmd                 RequestCommand = "CreateOrder"
	ActivateOrderCmd               RequestCommand = "ActivateOrder"
	UpdateOrderCmd                 RequestCommand = "UpdateOrder"
	DeleteOrderCmd                 RequestCommand = "DeleteOrder"
	AccountPositionsCmd            RequestCommand = "AccountPositions"
	AccountTradesCmd               RequestCommand = "AccountTrades"
	CountriesCmd                   RequestCommand = "Countries"
	IndicatorsCmd                  RequestCommand = "Indicators"
	InstrumentsCmd                 RequestCommand = "Instruments"
	InstrumentSearchCmd            RequestCommand = "InstrumentSearch"
	InstrumentLookupCmd            RequestCommand = "InstrumentLookup"
	InstrumentLeveragesCmd         RequestCommand = "InstrumentLeverages"
	InstrumentLeverageFiltersCmd   RequestCommand = "InstrumentLeverageFilters"
	InstrumentOptionPairsCmd       RequestCommand = "InstrumentOptionPairs"
	InstrumentOptionPairFiltersCmd RequestCommand = "InstrumentOptionPairFilters"
	InstrumentSectorsCmd           RequestCommand = "InstrumentSectors"
	InstrumentSectorCmd            RequestCommand = "InstrumentSector"
	InstrumentTypesCmd             RequestCommand = "InstrumentTypes"
	InstrumentUnderlyingsCmd       RequestCommand = "InstrumentUnderlyings"
	ListsCmd                       RequestCommand = "Lists"
	ListCmd                        RequestCommand = "List"
	MarketCmd                      RequestCommand = "Market"
	SearchNewsCmd                  RequestCommand = "SearchNews"
	NewsCmd                        RequestCommand = "News"
	NewsSourcesCmd                 RequestCommand = "NewsSources"
	RealtimeAccessCmd              RequestCommand = "RealtimeAccess"
	TickSizesCmd                   RequestCommand = "TickSizes"
	TickSizeCmd                    RequestCommand = "TickSize"
	TradableInfoCmd                RequestCommand = "TradableInfo"
	TradableIntradayCmd            RequestCommand = "TradableIntraday"
	TradableTradesCmd              RequestCommand = "TradableTrades"

	FeedSubCmd    RequestCommand = "FeedSubscribe"
	FeedUnsubCmd  RequestCommand = "FeedUnsubscribe"
	FeedStatusCmd RequestCommand = "FeedStatus"

	FeedLastCmd RequestCommand = "FeedLast"
)

// Is used as return struct for TransportRespondsToCmd
type RequestCommandInfo struct {
	Command    RequestCommand                        `json:"cmd"`
	Desc       string                                `json:"desc,omitempty"`
	Arguments  []RequestArgumentInfo                 `json:"args,omitempty"`
	TimeToLive int64                                 `json:"ttl,omitempty"`
	HandlerFn  func(Params) (json.RawMessage, error) `json:"-"` // If implemented, we can
}

// For builder pattern
func (rci *RequestCommandInfo) Description(desc string) *RequestCommandInfo {
	rci.Desc = desc
	return rci
}

// For builder pattern
func (rci *RequestCommandInfo) TTLHours(hours int64) *RequestCommandInfo {
	return rci.TTL(hours * 60 * 60 * 1000)
}

// For builder pattern
func (rci *RequestCommandInfo) TTL(ttl int64) *RequestCommandInfo {
	rci.TimeToLive = ttl
	return rci
}

// For builder pattern
func (rci *RequestCommandInfo) Handler(fun func(Params) (json.RawMessage, error)) *RequestCommandInfo {
	rci.HandlerFn = fun
	return rci
}

// For builder pattern
func (rci *RequestCommandInfo) AddArgument(name string) *RequestCommandInfo {
	return rci.AddFullArgument(name, "", []string{}, false)
}

// For builder pattern
func (rci *RequestCommandInfo) AddOptArgument(name string) *RequestCommandInfo {
	return rci.AddFullArgument(name, "", []string{}, true)
}

// For builder pattern
func (rci *RequestCommandInfo) AddFullArgument(name, desc string, opts []string, optional bool) *RequestCommandInfo {
	rci.Arguments = append(rci.Arguments, RequestArgumentInfo{
		Name:        name,
		Description: desc,
		Options:     opts,
		Optional:    optional,
	})
	return rci
}

func (rci *RequestCommandInfo) GetArgumentNames() (res []string) {
	for _, arg := range rci.Arguments {
		res = append(res, arg.Name)
	}
	return
}

type RequestArgumentInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"desc,omitempty"`
	Options     []string `json:"opts,omitempty"`
	Optional    bool     `json:"optional,omitempty"`

	Address string `json:"address,omitempty"` // Optional.  Use when the command is located on none standard topic
}

type RequestCommandTransport map[RequestCommand]*RequestCommandInfo

func (rct RequestCommandTransport) AddCommand(name string) *RequestCommandInfo {
	rci := &RequestCommandInfo{Command: RequestCommand(name)} // Make it now, and add it...
	rct[RequestCommand(name)] = rci
	return rci
}

func (rct RequestCommandTransport) Preform(req *Request) (res Response) {
	if req.Command == TransportRespondsToCmd {
		arr := []*RequestCommandInfo{}
		for _, rci := range rct {
			arr = append(arr, rci)
		}
		res.Success(arr)
	} else if cmd, ok := rct[req.Command]; ok {
		if cmd.HandlerFn != nil {
			r, err := cmd.HandlerFn(req.Args)
			if err != nil {
				res.Fail(-16, err.Error())
			} else {
				res.Payload = r
			}
		} else {
			res.Fail(-17, "Command does not have a method to be executed")
		}
	} else {
		res.Fail(-18, "Command not found")
	}
	return
}

type Params map[string]string

func (p Params) Pluck(keys ...string) (res []string) {
	for _, k := range keys {
		res = append(res, p[k])
	}
	return
}

func (p Params) Sprintf(format string, keys ...string) string {
	target := []interface{}{}
	for _, k := range keys {
		target = append(target, p[k])
	}
	return fmt.Sprintf(format, target...)
}

func (p Params) SubParams(keys ...string) (res Params) {
	res = make(Params)
	for _, k := range keys {
		if val, ok := p[k]; ok {
			res[k] = val
		}
	}
	return
}

type Request struct {
	Command RequestCommand `json:"cmd"`
	Args    Params         `json:"args,omitempty"`
}

func NewRequest(command RequestCommand, params map[string]string) (req *Request, err error) {
	req = &Request{Command: command, Args: Params{}}
	if params != nil {
		for k, v := range params {
			req.Args[k] = v
		}
	}

	fmt.Printf("Request %s -> %s\n", req.Command, req.Args)
	return
}

func (req *Request) Encode() ([]byte, error) {
	return json.Marshal(req)
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

type infoAwareTransportHandler struct {
	TransportHandler
	RequestCommandInfo
}

//
type TransportCacheHandler interface {
	Handle(RequestCommandInfo, TransportHandler, *Request) Response
}
type TransportCacheHandlerFn func(RequestCommandInfo, TransportHandler, *Request) Response

func (tchf TransportCacheHandlerFn) Handle(rci RequestCommandInfo, th TransportHandler, r *Request) Response {
	return tchf(rci, th, r)
}

type TransportRouter struct {
	routed       map[RequestCommand]infoAwareTransportHandler
	cacheHandler TransportCacheHandler
}

func NewTransportRouter(transports ...TransportHandler) (tr *TransportRouter, err error) {
	dummyCache := TransportCacheHandlerFn(func(rci RequestCommandInfo, th TransportHandler, r *Request) Response {
		return th.Preform(r)
	})
	return NewCachedTransportRouter(dummyCache, transports...)
}

func NewCachedTransportRouter(ch TransportCacheHandler, transports ...TransportHandler) (tr *TransportRouter, err error) {
	tr = &TransportRouter{
		routed:       make(map[RequestCommand]infoAwareTransportHandler),
		cacheHandler: ch,
	}
	for _, th := range transports {
		err = tr.AddTransportHandler(th)
		if err != nil {
			return
		}
	}
	return
}

func (tr *TransportRouter) AddTransportHandler(th TransportHandler) (err error) {
	var cmds []RequestCommandInfo
	res := th.Preform(&Request{Command: TransportRespondsToCmd, Args: Params{}})
	err = res.Unmarshal(&cmds)
	//	err = json.Unmarshal(res.Payload, &cmds)
	if err == nil {
		for _, cmd := range cmds {
			tr.routed[cmd.Command] = infoAwareTransportHandler{th, cmd}
		}
	} else {
		fmt.Printf("\nError unmarshal response of (%s): %+v\n\n", res.String(), err)
	}
	return
}

func (tr TransportRouter) Preform(req *Request) (res Response) {
	if req.Command == TransportRespondsToCmd {
		resArgs := []RequestCommandInfo{}
		for _, rci := range tr.routed {
			resArgs = append(resArgs, rci.RequestCommandInfo)
		}
		res.Success(resArgs)
		return
	}

	if iath, ok := tr.routed[req.Command]; ok {
		return tr.cacheHandler.Handle(iath.RequestCommandInfo, iath.TransportHandler, req)
	}
	cmds := []RequestCommand{}
	for cmd, _ := range tr.routed {
		cmds = append(cmds, cmd)
	}
	res.Fail(-1811, fmt.Sprintf("No command mapped for %+v: Available: %+v", req, cmds))
	return
}
