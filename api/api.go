package api

import (
	"encoding/json"
	"github.com/Forau/yanngo/swagger"

	//  "fmt"
	"strings"
)

type ApiClient struct {
	ph TransportHandler
}

func NewApiClient(ph TransportHandler) *ApiClient {
	return &ApiClient{ph: ph}
}

func (ac *ApiClient) handleRequest(command RequestCommand, query interface{}, res interface{}) error {
	req, err := NewRequest(command, query)
	if err != nil {
		return err
	}
	resp := ac.ph.Preform(req)
	if resp.Error != nil {
		return resp.Error
	} else {
		return json.Unmarshal(resp.Payload, res)
	}
}

type accountQuery struct {
	Accno   int64 `json:"accno,omitempty"`
	OrderId int64 `json:"orderid,omitempty"`
}

func (ac *ApiClient) Session() (res swagger.Login, err error) {
	err = ac.handleRequest(SessionCmd, nil, &res)
	return
}
func (ac *ApiClient) Accounts() (res []swagger.Account, err error) {
	err = ac.handleRequest(AccountsCmd, nil, &res)
	return
}
func (ac *ApiClient) Account(accno int64) (res swagger.AccountInfo, err error) {
	err = ac.handleRequest(AccountCmd, &accountQuery{Accno: accno}, &res)
	return
}
func (ac *ApiClient) AccountLedgers(accno int64) (res swagger.LedgerInformation, err error) {
	err = ac.handleRequest(AccountLedgersCmd, &accountQuery{Accno: accno}, &res)
	return
}
func (ac *ApiClient) AccountOrders(accno int64) (res swagger.Order, err error) {
	err = ac.handleRequest(AccountOrdersCmd, &accountQuery{Accno: accno}, &res)
	return
}
func (ac *ApiClient) CreateOrder(accno int64) (res swagger.OrderReply, err error) {
	err = ac.handleRequest(CreateOrderCmd, &accountQuery{Accno: accno}, &res)
	return
}
func (ac *ApiClient) ActivateOrder(accno int64, id int64) (res swagger.OrderReply, err error) {
	err = ac.handleRequest(ActivateOrderCmd, &accountQuery{Accno: accno, OrderId: id}, &res)
	return
}
func (ac *ApiClient) UpdateOrder(accno int64, id int64) (res swagger.OrderReply, err error) {
	err = ac.handleRequest(UpdateOrderCmd, &accountQuery{Accno: accno, OrderId: id}, &res)
	return
}
func (ac *ApiClient) DeleteOrder(accno int64, id int64) (res swagger.OrderReply, err error) {
	err = ac.handleRequest(DeleteOrderCmd, &accountQuery{Accno: accno, OrderId: id}, &res)
	return
}
func (ac *ApiClient) AccountPositions(accno int64) (res []swagger.Position, err error) {
	err = ac.handleRequest(AccountPositionsCmd, &accountQuery{Accno: accno}, &res)
	return
}
func (ac *ApiClient) AccountTrades(accno int64) (res []swagger.Trade, err error) {
	err = ac.handleRequest(AccountTradesCmd, &accountQuery{Accno: accno}, &res)
	return
}
func (ac *ApiClient) Countries(countries ...string) (res []swagger.Country, err error) {
	err = ac.handleRequest(CountriesCmd, map[string][]string{"countries": countries}, &res)
	return
}
func (ac *ApiClient) Indicators(indicators ...string) (res []swagger.Indicator, err error) {
	err = ac.handleRequest(IndicatorsCmd, map[string][]string{"indicators": indicators}, &res)
	return
}
func (ac *ApiClient) InstrumentSearch(query string) (res []swagger.Instrument, err error) {
	err = ac.handleRequest(InstrumentSearchCmd, map[string]string{"query": query}, &res)
	return
}
func (ac *ApiClient) Instruments(ids ...int64) (res []swagger.Instrument, err error) {
	err = ac.handleRequest(InstrumentsCmd, map[string][]int64{"ids": ids}, &res)
	return
}

/*
func (ac *ApiClient) InstrumentLeverages(pathMap{id int64})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentLeveragesCmd, &)
  return
}
func (ac *ApiClient) InstrumentLeverageFilters(pathMap{id int64})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentLeverageFiltersCmd, &)
  return
}
func (ac *ApiClient) InstrumentOptionPairs(pathMap{id int64})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentOptionPairsCmd, &)
  return
}
func (ac *ApiClient) InstrumentOptionPairFilters(pathMap{id int64})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentOptionPairFiltersCmd, &)
  return
}
*/
// The lookup_type is isin_code_currency_market_id or market_id_identifier
func (ac *ApiClient) InstrumentLookup(typ, instrument string) (res []swagger.Instrument, err error) {
	query := map[string]string{"lookup": instrument, "type": typ}
	err = ac.handleRequest(InstrumentLookupCmd, query, &res)
	return
}

/*
func (ac *ApiClient) InstrumentSectors()(res swagger.,err error) {
  err = ac.handleRequest(InstrumentSectorsCmd, &)
  return
}
func (ac *ApiClient) InstrumentSector(pathMap{pathMap{"Sectors",fmtInt}})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentSectorCmd, &)
  return
}
func (ac *ApiClient) InstrumentTypes()(res swagger.,err error) {
  err = ac.handleRequest(InstrumentTypesCmd, &)
  return
}
func (ac *ApiClient) InstrumentType(pathMap{pathMap{"Type",fmtInt}})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentTypeCmd, &)
  return
}
func (ac *ApiClient) InstrumentUnderlyings(pathMap{pathMap{"Type",fmtInt},pathMap{"Currency",fmtInt}})(res swagger.,err error) {
  err = ac.handleRequest(InstrumentUnderlyingsCmd, &)
  return
}
*/

func (ac *ApiClient) Lists() (res []swagger.List, err error) {
	err = ac.handleRequest(ListsCmd, nil, &res)
	return
}
func (ac *ApiClient) List(id int64) (res []swagger.Instrument, err error) {
	err = ac.handleRequest(ListCmd, map[string]int64{"id": id}, &res)
	return
}

func (ac *ApiClient) Market(ids ...int64) (res []swagger.Market, err error) {
	err = ac.handleRequest(MarketCmd, map[string][]int64{"ids": ids}, &res)
	return
}

/*
func (ac *ApiClient) SearchNews()(res swagger.,err error) {
  err = ac.handleRequest(SearchNewsCmd, &)
  return
}
func (ac *ApiClient) News(pathMap{ids []int64})(res swagger.,err error) {
  err = ac.handleRequest(NewsCmd, &)
  return
}
func (ac *ApiClient) NewsSources()(res swagger.,err error) {
  err = ac.handleRequest(NewsSourcesCmd, &)
  return
}
func (ac *ApiClient) RealtimeAccess()(res swagger.,err error) {
  err = ac.handleRequest(RealtimeAccessCmd, &)
  return
}
func (ac *ApiClient) TickSizes()(res swagger.,err error) {
  err = ac.handleRequest(TickSizesCmd, &)
  return
}
func (ac *ApiClient) TickSize(pathMap{ids []int64})(res swagger.,err error) {
  err = ac.handleRequest(TickSizeCmd, &)
  return
}
*/
func (ac *ApiClient) TradableInfo(ids ...string) (res []swagger.TradableInfo, err error) {
	err = ac.handleRequest(TradableInfoCmd, map[string]string{"ids": strings.Join(ids, ",")}, &res)
	return
}
func (ac *ApiClient) TradableDay(ids ...string) (res []swagger.IntradayGraph, err error) {
	err = ac.handleRequest(TradableIntradayCmd, map[string]string{"ids": strings.Join(ids, ",")}, &res)
	return
}
func (ac *ApiClient) TradableTrades(ids ...string) (res []swagger.PublicTrades, err error) {
	query := map[string]string{"ids": strings.Join(ids, ",")}
	/*  if count > 0 {
	      query["count"] = fmt.Sprintf("%d", count)
	    } else {
	      query["count"] = "all"
	    }
	*/
	err = ac.handleRequest(TradableTradesCmd, query, &res)
	return
}
