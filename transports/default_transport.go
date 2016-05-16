package transports

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/httpcli"

	"fmt"
)

type pathMap struct {
	key  string
	conv func(in interface{}) string
}

var fmtStr = func(in interface{}) string {
	if str, ok := in.(string); ok {
		return str
	}
	fmt.Printf("Dont know how to convert %T (%+v) to int.\n", in, in)
	return fmt.Sprintf("%v", in)
}

var fmtInt = func(in interface{}) string {
	switch t := in.(type) {
	case float64, float32:
		return fmt.Sprintf("%.0f", t)
	case int, uint:
		return fmt.Sprintf("%d", t)
	default:
		fmt.Printf("Dont know how to convert %T (%+v) to int.\n", t, t)
		return fmt.Sprintf("%v", in)
	}
}

var fmtIntArr = func(in interface{}) string {
	if arr, ok := in.([]interface{}); ok {
		res := []byte{}
		for idx, val := range arr {
			if idx > 0 {
				res = append(res, ',')
			}
			res = append(res, fmtInt(val)...)
		}
		return string(res)
	} else {
		return fmtInt(in)
	}
}

type commandMapper struct {
	action    string
	path      string
	pathParts []pathMap
}

// For now we use template, but later we might add logic to have a static field to save resources
func (cm *commandMapper) execute(in map[string]interface{}) string {
	if len(cm.pathParts) > 0 {
		vals := []interface{}{}
		for _, k := range cm.pathParts {
			vals = append(vals, k.conv(in[k.key]))
			delete(in, k.key)
		}
		return fmt.Sprintf(cm.path, vals...)
	}
	return cm.path
}

func newCommandMapper(action, path string, pathParts []pathMap) *commandMapper {
	return &commandMapper{
		action:    action,
		path:      path,
		pathParts: pathParts,
	}
}

var CommandTemplates = map[api.RequestCommand]*commandMapper{
	api.SystemStatusCmd:                newCommandMapper("GET", "", []pathMap{}),
	api.LoginCmd:                       newCommandMapper("POST", "login", []pathMap{}),
	api.LogoutCmd:                      newCommandMapper("DELETE", "login", []pathMap{}),
	api.TouchCmd:                       newCommandMapper("PUT", "login", []pathMap{}),
	api.AccountsCmd:                    newCommandMapper("GET", "accounts", []pathMap{}),
	api.AccountCmd:                     newCommandMapper("GET", "accounts/%s", []pathMap{pathMap{"Accno", fmtInt}}),
	api.AccountLedgersCmd:              newCommandMapper("GET", "accounts/%s/ledgers", []pathMap{pathMap{"Accno", fmtInt}}),
	api.AccountOrdersCmd:               newCommandMapper("GET", "accounts/%s/orders", []pathMap{pathMap{"Accno", fmtInt}}),
	api.CreateOrderCmd:                 newCommandMapper("POST", "accounts/%s/orders", []pathMap{pathMap{"Accno", fmtInt}}),
	api.ActivateOrderCmd:               newCommandMapper("PUT", "accounts/%s/orders/%s/activate", []pathMap{pathMap{"Accno", fmtInt}, pathMap{"Id", fmtInt}}),
	api.UpdateOrderCmd:                 newCommandMapper("PUT", "accounts/%s/orders/%s", []pathMap{pathMap{"Accno", fmtInt}, pathMap{"Id", fmtInt}}),
	api.DeleteOrderCmd:                 newCommandMapper("DELETE", "accounts/%s/orders/%s", []pathMap{pathMap{"Accno", fmtInt}, pathMap{"Id", fmtInt}}),
	api.AccountPositionsCmd:            newCommandMapper("GET", "accounts/%s/positions", []pathMap{pathMap{"Accno", fmtInt}}),
	api.AccountTradesCmd:               newCommandMapper("GET", "accounts/%s/trades", []pathMap{pathMap{"Accno", fmtInt}}),
	api.CountriesCmd:                   newCommandMapper("GET", "countries", []pathMap{}),
	api.LookupCountriesCmd:             newCommandMapper("GET", "countries/%s", []pathMap{pathMap{"Countries", fmtInt}}),
	api.IndicatorsCmd:                  newCommandMapper("GET", "indicators", []pathMap{}),
	api.LookupIndicatorsCmd:            newCommandMapper("GET", "indicators/%s", []pathMap{pathMap{"Indicators", fmtInt}}),
	api.SearchInstrumentsCmd:           newCommandMapper("GET", "instruments", []pathMap{}),
	api.InstrumentsCmd:                 newCommandMapper("GET", "instruments/%s", []pathMap{pathMap{"Ids", fmtIntArr}}),
	api.InstrumentLeveragesCmd:         newCommandMapper("GET", "instruments/%s/leverages", []pathMap{pathMap{"Id", fmtInt}}),
	api.InstrumentLeverageFiltersCmd:   newCommandMapper("GET", "instruments/%s/leverages/filters", []pathMap{pathMap{"Id", fmtInt}}),
	api.InstrumentOptionPairsCmd:       newCommandMapper("GET", "instruments/%s/option_pairs", []pathMap{pathMap{"Id", fmtInt}}),
	api.InstrumentOptionPairFiltersCmd: newCommandMapper("GET", "instruments/%s/option_pairs/filters", []pathMap{pathMap{"Id", fmtInt}}),
	api.InstrumentLookupCmd:            newCommandMapper("GET", "instruments/lookup/%s/%s", []pathMap{pathMap{"Type", fmtInt}, pathMap{"Lookup", fmtInt}}),
	api.InstrumentSectorsCmd:           newCommandMapper("GET", "instruments/sectors", []pathMap{}),
	api.InstrumentSectorCmd:            newCommandMapper("GET", "instruments/sectors/%s", []pathMap{pathMap{"Sectors", fmtInt}}),
	api.InstrumentTypesCmd:             newCommandMapper("GET", "instruments/types", []pathMap{}),
	api.InstrumentTypeCmd:              newCommandMapper("GET", "instruments/types/%s", []pathMap{pathMap{"Type", fmtInt}}),
	api.InstrumentUnderlyingsCmd:       newCommandMapper("GET", "instruments/underlyings/%s/%s", []pathMap{pathMap{"Type", fmtInt}, pathMap{"Currency", fmtInt}}),
	api.ListsCmd:                       newCommandMapper("GET", "lists", []pathMap{}),
	api.ListCmd:                        newCommandMapper("GET", "lists/%s", []pathMap{pathMap{"Id", fmtInt}}),
	api.MarketsCmd:                     newCommandMapper("GET", "markets", []pathMap{}),
	api.MarketCmd:                      newCommandMapper("GET", "markets/%s", []pathMap{pathMap{"Ids", fmtInt}}),
	api.SearchNewsCmd:                  newCommandMapper("GET", "news", []pathMap{}),
	api.NewsCmd:                        newCommandMapper("GET", "news/%s", []pathMap{pathMap{"Ids", fmtInt}}),
	api.NewsSourcesCmd:                 newCommandMapper("GET", "news_sources", []pathMap{}),
	api.RealtimeAccessCmd:              newCommandMapper("GET", "realtime_access", []pathMap{}),
	api.TickSizesCmd:                   newCommandMapper("GET", "tick_sizes", []pathMap{}),
	api.TickSizeCmd:                    newCommandMapper("GET", "tick_sizes/%s", []pathMap{pathMap{"Ids", fmtInt}}),
	api.TradableInfoCmd:                newCommandMapper("GET", "tradables/info/%s", []pathMap{pathMap{"Ids", fmtInt}}),
	api.TradableIntradayCmd:            newCommandMapper("GET", "tradables/intraday/%s", []pathMap{pathMap{"Ids", fmtInt}}),
	api.TradableTradesCmd:              newCommandMapper("GET", "tradables/trades/%s", []pathMap{pathMap{"Ids", fmtInt}}),
}

func NewDefaultTransport(endpoint string, user, pass, rawPem []byte) (transp api.Transport, err error) {
	restcli := httpcli.NewRestClient(endpoint, user, pass, rawPem)

	transp = func(req *api.Request) (res api.Response) {
		defer func() {
			if err := recover(); err != nil {
				res.Fail(-1, fmt.Sprintf("%+v", err))
			}
		}()

		if templ := CommandTemplates[req.Command]; templ != nil {
			qmap, err := req.QueryMap()
			if err != nil {
				fmt.Printf("Error getting query map [%+v]: %+v\n", req, err)
				panic(err) // TODO: fix
			}
			path := templ.execute(qmap)

			fmt.Printf("Got my path: %+v\n", path)
			resp, err := restcli.Execute(templ.action, path, qmap)
			fmt.Printf("Res : %+v -> %+v\n", res, err)
			if err != nil {
				res.Fail(-2, err.Error())
			} else {
				res.Payload = resp
			}
		} else {
			panic(fmt.Errorf("Unable to find template for %s", req.Command))
		}
		return
	}
	return
}
