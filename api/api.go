// Copyright (c) 2016 Forau @ github.com. MIT License.

// Package api contains definitions for the client and transports
package api

import (
	"encoding/json"
	"github.com/Forau/yanngo/swagger"

	"fmt"
	"strings"
)

type requestBuilder struct {
	req *Request
	err error
	ph  TransportHandler
}

func (rb *requestBuilder) Exec(res interface{}) error {
	if rb.err != nil {
		return rb.err
	}
	resp := rb.ph.Preform(rb.req)
	if resp.Error != nil {
		return resp.Error
	} else {
		return json.Unmarshal(resp.Payload, res)
	}
}

func (rb *requestBuilder) S(key, val string) *requestBuilder {
	rb.req.Params[key] = val
	return rb
}
func (rb *requestBuilder) I(key string, val uint64) *requestBuilder {
	return rb.S(key, fmt.Sprintf("%d", val))
}
func (rb *requestBuilder) IA(key string, val []uint64) *requestBuilder {
	valArr := []string{}
	for _, v := range val {
		valArr = append(valArr, fmt.Sprintf("%d", v))
	}
	return rb.S(key, strings.Join(valArr, ","))
}
func (rb *requestBuilder) F(key string, val float64) *requestBuilder {
	return rb.S(key, fmt.Sprintf("%f", val))
}
func (rb *requestBuilder) FA(key string, val []float64) *requestBuilder {
	valArr := []string{}
	for _, v := range val {
		valArr = append(valArr, fmt.Sprintf("%f", v))
	}
	return rb.S(key, strings.Join(valArr, ","))
}

type ApiClient struct {
	ph TransportHandler
}

func NewApiClient(ph TransportHandler) *ApiClient {
	return &ApiClient{ph: ph}
}

func (ac *ApiClient) build(command RequestCommand) *requestBuilder {
	return &requestBuilder{
		req: &Request{Command: command, Params: Params{}},
		ph:  ac.ph,
	}
}

type accountQuery struct {
	Accno   uint64 `json:"accno,omitempty"`
	OrderId uint64 `json:"orderid,omitempty"`
}

type OrderSide string

func (os OrderSide) Buy() OrderSide  { return "BUY" }
func (os OrderSide) Sell() OrderSide { return "SELL" }

type OrderType string

const (
	FAK           OrderType = "FAK"
	FOK                     = "FOK"
	LIMIT                   = "LIMIT"
	STOP_LIMIT              = "STOP_LIMIT"
	STOP_TRAILING           = "STOP_TRAILING"
	OCO                     = "OCO"
)

type OrderCondition string
type OrderTriggerDir string

type AccountOrder struct {
	Accno   uint64 `json:"accno"`
	OrderId uint64 `json:"orderid,omitempty"` // Only when modifying existing

	// Mandatory
	Identifier string    `json:"identifier"`
	MarketId   uint64    `json:"market_id"`
	Price      float64   `json:"price"`
	Volume     uint64    `json:"volume,omitempty"`
	Side       OrderSide `json:"side"`

	// Omited == LIMIT
	OrderType OrderType `json:"order_type,omitempty"`

	// Special
	Currency            string          `json:"currency,omitempty"`
	ValidUntil          string          `json:"valid_until,omitempty"`
	OpenVolume          uint64          `json:"open_volume,omitempty"`
	Reference           string          `json:"reference,omitempty"`
	ActivationCondition OrderCondition  `json:"activation_condition,omitempty"`
	TriggerCondition    OrderTriggerDir `json:"trigger_condition,omitempty"`
	TargetValue         float64         `json:"target_value,omitempty"`
}

func (ac *ApiClient) Session() (res swagger.Login, err error) {
	err = ac.build(SessionCmd).Exec(&res)
	return
}
func (ac *ApiClient) Accounts() (res []swagger.Account, err error) {
	err = ac.build(AccountsCmd).Exec(&res)
	return
}
func (ac *ApiClient) Account(accno uint64) (res swagger.AccountInfo, err error) {
	err = ac.build(AccountCmd).I("accno", accno).Exec(&res)
	return
}
func (ac *ApiClient) AccountLedgers(accno uint64) (res swagger.LedgerInformation, err error) {
	err = ac.build(AccountLedgersCmd).I("accno", accno).Exec(&res)
	return
}
func (ac *ApiClient) AccountOrders(accno uint64) (res []swagger.Order, err error) {
	err = ac.build(AccountOrdersCmd).I("accno", accno).Exec(&res)
	return
}
func (ac *ApiClient) CreateSimpleOrder(accno uint64, identifier string, market uint64,
	price float64, vol uint64, side string) (res swagger.OrderReply, err error) {
	err = ac.build(CreateOrderCmd).I("accno", accno).S("identifier", identifier).I("market_id", market).
		F("price", price).I("volume", vol).S("side", side).
		S("currency", "SEK").Exec(&res)
	return
}
func (ac *ApiClient) CreateOrder(order *AccountOrder) (res swagger.OrderReply, err error) {
	err = ac.build(CreateOrderCmd).Exec(&res) // TODO
	return
}

func (ac *ApiClient) ActivateOrder(accno, id uint64) (res swagger.OrderReply, err error) {
	err = ac.build(ActivateOrderCmd).I("accno", accno).I("order_id", id).Exec(&res)
	return
}
func (ac *ApiClient) UpdateOrder(accno, id uint64) (res swagger.OrderReply, err error) {
	err = ac.build(UpdateOrderCmd).I("accno", accno).I("order_id", id).Exec(&res)
	return
}
func (ac *ApiClient) DeleteOrder(accno, id uint64) (res swagger.OrderReply, err error) {
	err = ac.build(DeleteOrderCmd).I("accno", accno).I("order_id", id).Exec(&res)
	return
}
func (ac *ApiClient) AccountPositions(accno uint64) (res []swagger.Position, err error) {
	err = ac.build(AccountPositionsCmd).I("accno", accno).Exec(&res)
	return
}
func (ac *ApiClient) AccountTrades(accno uint64) (res []swagger.Trade, err error) {
	err = ac.build(AccountTradesCmd).I("accno", accno).Exec(&res)
	return
}
func (ac *ApiClient) Countries(countries ...string) (res []swagger.Country, err error) {
	err = ac.build(CountriesCmd).S("countries", strings.Join(countries, ",")).Exec(&res)
	return
}
func (ac *ApiClient) Indicators(indicators ...string) (res []swagger.Indicator, err error) {
	err = ac.build(IndicatorsCmd).S("indicators", strings.Join(indicators, ",")).Exec(&res)
	return
}
func (ac *ApiClient) InstrumentSearch(query string) (res []swagger.Instrument, err error) {
	err = ac.build(InstrumentSearchCmd).S("query", query).Exec(&res)
	return
}
func (ac *ApiClient) Instruments(ids ...uint64) (res []swagger.Instrument, err error) {
	err = ac.build(InstrumentsCmd).IA("ids", ids).Exec(&res)
	return
}

/*
func (ac *ApiClient) InstrumentLeverages(pathMap{id uint64})(res swagger.,err error) {
  err = ac.build(InstrumentLeveragesCmd). &)
  return
}
func (ac *ApiClient) InstrumentLeverageFilters(pathMap{id uint64})(res swagger.,err error) {
  err = ac.build(InstrumentLeverageFiltersCmd). &)
  return
}
func (ac *ApiClient) InstrumentOptionPairs(pathMap{id uint64})(res swagger.,err error) {
  err = ac.build(InstrumentOptionPairsCmd). &)
  return
}
func (ac *ApiClient) InstrumentOptionPairFilters(pathMap{id uint64})(res swagger.,err error) {
  err = ac.build(InstrumentOptionPairFiltersCmd). &)
  return
}
*/
// The lookup_type is isin_code_currency_market_id or market_id_identifier
func (ac *ApiClient) InstrumentLookup(typ, instrument string) (res []swagger.Instrument, err error) {
	err = ac.build(InstrumentLookupCmd).S("lookup", instrument).S("type", typ).Exec(&res)
	return
}

/*
func (ac *ApiClient) InstrumentSectors()(res swagger.,err error) {
  err = ac.build(InstrumentSectorsCmd). &)
  return
}
func (ac *ApiClient) InstrumentSector(pathMap{pathMap{"Sectors",fmtInt}})(res swagger.,err error) {
  err = ac.build(InstrumentSectorCmd). &)
  return
}
func (ac *ApiClient) InstrumentTypes()(res swagger.,err error) {
  err = ac.build(InstrumentTypesCmd). &)
  return
}
func (ac *ApiClient) InstrumentType(pathMap{pathMap{"Type",fmtInt}})(res swagger.,err error) {
  err = ac.build(InstrumentTypeCmd). &)
  return
}
func (ac *ApiClient) InstrumentUnderlyings(pathMap{pathMap{"Type",fmtInt},pathMap{"Currency",fmtInt}})(res swagger.,err error) {
  err = ac.build(InstrumentUnderlyingsCmd). &)
  return
}
*/

func (ac *ApiClient) Lists() (res []swagger.List, err error) {
	err = ac.build(ListsCmd).Exec(&res)
	return
}
func (ac *ApiClient) List(id uint64) (res []swagger.Instrument, err error) {
	err = ac.build(ListCmd).I("id", id).Exec(&res)
	return
}

func (ac *ApiClient) Market(ids ...uint64) (res []swagger.Market, err error) {
	err = ac.build(MarketCmd).IA("ids", ids).Exec(&res)
	return
}

/*
func (ac *ApiClient) SearchNews()(res swagger.,err error) {
  err = ac.build(SearchNewsCmd). &)
  return
}
func (ac *ApiClient) News(pathMap{ids []uint64})(res swagger.,err error) {
  err = ac.build(NewsCmd). &)
  return
}
func (ac *ApiClient) NewsSources()(res swagger.,err error) {
  err = ac.build(NewsSourcesCmd). &)
  return
}
*/

func (ac *ApiClient) RealtimeAccess() (res []swagger.RealtimeAccess, err error) {
	err = ac.build(RealtimeAccessCmd).Exec(&res)
	return
}
func (ac *ApiClient) TickSizes() (res []swagger.TicksizeTable, err error) {
	err = ac.build(TickSizesCmd).Exec(&res)
	return
}
func (ac *ApiClient) TickSize(ids ...uint64) (res []swagger.TicksizeTable, err error) {
	err = ac.build(TickSizeCmd).IA("ids", ids).Exec(&res)
	return
}

func (ac *ApiClient) TradableInfo(ids ...string) (res []swagger.TradableInfo, err error) {
	err = ac.build(TradableInfoCmd).S("ids", strings.Join(ids, ",")).Exec(&res)
	return
}
func (ac *ApiClient) TradableDay(ids ...string) (res []swagger.IntradayGraph, err error) {
	err = ac.build(TradableIntradayCmd).S("ids", strings.Join(ids, ",")).Exec(&res)
	return
}
func (ac *ApiClient) TradableTrades(ids ...string) (res []swagger.PublicTrades, err error) {
	err = ac.build(TradableTradesCmd).S("ids", strings.Join(ids, ",")).Exec(&res)
	return
}

// Feeds
func (ac *ApiClient) FeedSub(typ, id, mark string) (res map[string]string, err error) {
	err = ac.build(FeedSubCmd).S("type", typ).S("id", id).S("market", mark).Exec(&res)
	return
}

func (ac *ApiClient) FeedUnsub(subId string) (res map[string]string, err error) {
	err = ac.build(FeedUnsubCmd).S("id", subId).Exec(&res)
	return
}

func (ac *ApiClient) FeedStatus() (res map[string]interface{}, err error) {
	err = ac.build(FeedStatusCmd).Exec(&res)
	return
}
