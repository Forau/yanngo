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

type RequestCommand string

// The commands that we support.  Will be mapped in the low level TransportHandler
const (
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
)

type Request struct {
	Command RequestCommand
	Query   json.RawMessage // Will get unmashaled and converted to query or form params
}

func (r *Request) QueryMap() (res map[string]interface{}, err error) {
	if r.Query != nil {
		err = json.Unmarshal(r.Query, &res)
	}
	return
}

func NewRequest(command RequestCommand, query interface{}) (req *Request, err error) {
	req = &Request{Command: command}
	if query != nil {
		req.Query, err = json.Marshal(query)
	}
	fmt.Printf("Request %s -> %s\n", req.Command, string(req.Query))
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
