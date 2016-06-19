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

func (ar *Response) String() string {
	return fmt.Sprintf("Response{Error: %+v, Payload: %s}", ar.Error, string(ar.Payload))
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
)

type Params map[string]string

type Request struct {
	Command RequestCommand
	Params  Params
}

func NewRequest(command RequestCommand, params map[string]string) (req *Request, err error) {
	req = &Request{Command: command, Params: Params{}}
	if params != nil {
		for k, v := range params {
			req.Params[k] = v
		}
	}

	fmt.Printf("Request %s -> %s\n", req.Command, req.Params)
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

type TransportRouter struct {
	defTransport TransportHandler
	routed       map[RequestCommand]TransportHandler
}

func NewTransportRouter(defaultTr TransportHandler, others ...TransportHandler) (tr *TransportRouter, err error) {
	tr = &TransportRouter{defTransport: defaultTr, routed: make(map[RequestCommand]TransportHandler)}
	for _, th := range others {
		err = tr.AddTransportHandler(th)
		if err != nil {
			return
		}
	}
	return
}

func (tr *TransportRouter) AddTransportHandler(th TransportHandler) (err error) {
	var cmds []RequestCommand
	res := th.Preform(&Request{Command: TransportRespondsToCmd, Params: Params{}})
	err = json.Unmarshal(res.Payload, &cmds)
	if err == nil {
		for _, cmd := range cmds {
			tr.routed[cmd] = th
		}
	}
	return
}

func (tr TransportRouter) Preform(req *Request) Response {
	if th, ok := tr.routed[req.Command]; ok {
		return th.Preform(req)
	}
	return tr.defTransport.Preform(req)
}
