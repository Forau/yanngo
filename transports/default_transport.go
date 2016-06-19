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
	fmt.Printf("Dont know how to convert %T (%+v) to string.\n", in, in)
	return fmt.Sprintf("%v", in)
}

var fmtStrArr = func(in interface{}) string {
	fmt.Printf("Got possible array %+v\n", in)
	if arr, ok := in.([]interface{}); ok {
		res := []byte{}
		for idx, val := range arr {
			if idx > 0 {
				res = append(res, ',')
			}
			res = append(res, fmtStr(val)...)
		}
		return string(res)
	} else {
		return fmtStr(in)
	}
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
		fmt.Printf("Could not convert %+v to interface array.  %T\n", in, in)
		return fmtInt(in)
	}
}

type commandMapper struct {
	action    string
	path      string
	pathParts []pathMap
}

// For now we use template, but later we might add logic to have a static field to save resources
func (cm *commandMapper) execute(in map[string]string) string {
	if len(cm.pathParts) > 0 {
		vals := []interface{}{}
		for _, k := range cm.pathParts {
			vals = append(vals, in[k.key])
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
	api.SessionCmd:                     newCommandMapper("SPECIAL", "session", []pathMap{}),
	api.AccountsCmd:                    newCommandMapper("GET", "accounts", []pathMap{}),
	api.AccountCmd:                     newCommandMapper("GET", "accounts/%s", []pathMap{pathMap{"accno", fmtInt}}),
	api.AccountLedgersCmd:              newCommandMapper("GET", "accounts/%s/ledgers", []pathMap{pathMap{"accno", fmtInt}}),
	api.AccountOrdersCmd:               newCommandMapper("GET", "accounts/%s/orders", []pathMap{pathMap{"accno", fmtInt}}),
	api.CreateOrderCmd:                 newCommandMapper("POST", "accounts/%s/orders", []pathMap{pathMap{"accno", fmtInt}}),
	api.ActivateOrderCmd:               newCommandMapper("PUT", "accounts/%s/orders/%s/activate", []pathMap{pathMap{"accno", fmtInt}, pathMap{"order_id", fmtInt}}),
	api.UpdateOrderCmd:                 newCommandMapper("PUT", "accounts/%s/orders/%s", []pathMap{pathMap{"accno", fmtInt}, pathMap{"order_id", fmtInt}}),
	api.DeleteOrderCmd:                 newCommandMapper("DELETE", "accounts/%s/orders/%s", []pathMap{pathMap{"accno", fmtInt}, pathMap{"order_id", fmtInt}}),
	api.AccountPositionsCmd:            newCommandMapper("GET", "accounts/%s/positions", []pathMap{pathMap{"accno", fmtInt}}),
	api.AccountTradesCmd:               newCommandMapper("GET", "accounts/%s/trades", []pathMap{pathMap{"accno", fmtInt}}),
	api.CountriesCmd:                   newCommandMapper("GET", "countries/%s", []pathMap{pathMap{"countries", fmtStrArr}}),
	api.IndicatorsCmd:                  newCommandMapper("GET", "indicators/%s", []pathMap{pathMap{"indicators", fmtStrArr}}),
	api.InstrumentsCmd:                 newCommandMapper("GET", "instruments/%s", []pathMap{pathMap{"ids", fmtIntArr}}),
	api.InstrumentSearchCmd:            newCommandMapper("GET", "instruments", []pathMap{}),
	api.InstrumentLeveragesCmd:         newCommandMapper("GET", "instruments/%s/leverages", []pathMap{pathMap{"id", fmtInt}}),
	api.InstrumentLeverageFiltersCmd:   newCommandMapper("GET", "instruments/%s/leverages/filters", []pathMap{pathMap{"id", fmtInt}}),
	api.InstrumentOptionPairsCmd:       newCommandMapper("GET", "instruments/%s/option_pairs", []pathMap{pathMap{"id", fmtInt}}),
	api.InstrumentOptionPairFiltersCmd: newCommandMapper("GET", "instruments/%s/option_pairs/filters", []pathMap{pathMap{"id", fmtInt}}),
	api.InstrumentLookupCmd:            newCommandMapper("GET", "instruments/lookup/%s/%s", []pathMap{pathMap{"type", fmtStr}, pathMap{"lookup", fmtStr}}),
	api.InstrumentSectorsCmd:           newCommandMapper("GET", "instruments/sectors", []pathMap{}),
	api.InstrumentSectorCmd:            newCommandMapper("GET", "instruments/sectors/%s", []pathMap{pathMap{"sectors", fmtInt}}),
	api.InstrumentTypesCmd:             newCommandMapper("GET", "instruments/types/%s", []pathMap{pathMap{"type", fmtInt}}),
	api.InstrumentUnderlyingsCmd:       newCommandMapper("GET", "instruments/underlyings/%s/%s", []pathMap{pathMap{"type", fmtInt}, pathMap{"currency", fmtInt}}),
	api.ListsCmd:                       newCommandMapper("GET", "lists", []pathMap{}),
	api.ListCmd:                        newCommandMapper("GET", "lists/%s", []pathMap{pathMap{"id", fmtInt}}),
	api.MarketCmd:                      newCommandMapper("GET", "markets/%s", []pathMap{pathMap{"ids", fmtIntArr}}),
	api.SearchNewsCmd:                  newCommandMapper("GET", "news", []pathMap{}),
	api.NewsCmd:                        newCommandMapper("GET", "news/%s", []pathMap{pathMap{"ids", fmtInt}}),
	api.NewsSourcesCmd:                 newCommandMapper("GET", "news_sources", []pathMap{}),
	api.RealtimeAccessCmd:              newCommandMapper("GET", "realtime_access", []pathMap{}),
	api.TickSizesCmd:                   newCommandMapper("GET", "tick_sizes", []pathMap{}),
	api.TickSizeCmd:                    newCommandMapper("GET", "tick_sizes/%s", []pathMap{pathMap{"ids", fmtInt}}),
	api.TradableInfoCmd:                newCommandMapper("GET", "tradables/info/%s", []pathMap{pathMap{"ids", fmtInt}}),
	api.TradableIntradayCmd:            newCommandMapper("GET", "tradables/intraday/%s", []pathMap{pathMap{"ids", fmtInt}}),
	api.TradableTradesCmd:              newCommandMapper("GET", "tradables/trades/%s", []pathMap{pathMap{"ids", fmtInt}}),
}

func NewDefaultTransport(endpoint string, user, pass, rawPem []byte) (transp api.Transport, err error) {
	restcli := httpcli.NewRestClient(endpoint, user, pass, rawPem)

	transp = func(req *api.Request) (res api.Response) {
		defer func() {
			if err := recover(); err != nil {
				res.Fail(-1, fmt.Sprintf("%+v", err))
			}
		}()

		if req.Command == api.TransportRespondsToCmd {
			cmds := []api.RequestCommand{}
			for cmd, _ := range CommandTemplates {
				cmds = append(cmds, cmd)
			}
			res.Success(cmds)
		} else if templ := CommandTemplates[req.Command]; templ != nil {
			qmap := req.Params
			path := templ.execute(qmap)

			fmt.Printf("Got my path: %+v, and data %+v\n", path, qmap)
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
