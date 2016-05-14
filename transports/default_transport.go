package transports

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/crypto"
	"text/template"

	"bytes"
	"fmt"
)

type commandMapper struct {
	action string
	templ  *template.Template
}

// For now we use template, but later we might add logic to have a static field to save resources
func (cm *commandMapper) execute(in interface{}) (res string, err error) {
	var buf bytes.Buffer
	err = cm.templ.Execute(&buf, in)
	if err == nil {
		res = buf.String()
	}
	return
}

func newCommandMapper(action, templ string) *commandMapper {
	return &commandMapper{
		action: action,
		templ:  template.Must(template.New(action + templ).Parse(templ)),
	}
}

var CommandTemplates = map[api.RequestCommand]*commandMapper{
	api.SystemStatusCmd:                newCommandMapper("GET", ""),
	api.LoginCmd:                       newCommandMapper("POST", "login"),
	api.LogoutCmd:                      newCommandMapper("DELETE", "login"),
	api.TouchCmd:                       newCommandMapper("PUT", "login"),
	api.AccountsCmd:                    newCommandMapper("GET", "accounts"),
	api.AccountCmd:                     newCommandMapper("GET", "accounts/{{.Accno}}"),
	api.AccountLedgersCmd:              newCommandMapper("GET", "accounts/{{.Accno}}/ledgers"),
	api.AccountOrdersCmd:               newCommandMapper("GET", "accounts/{{.Accno}}/orders"),
	api.CreateOrderCmd:                 newCommandMapper("POST", "accounts/{{.Accno}}/orders"),
	api.ActivateOrderCmd:               newCommandMapper("PUT", "accounts/{{.Accno}}/orders/{{.Id}}/activate"),
	api.UpdateOrderCmd:                 newCommandMapper("PUT", "accounts/{{.Accno}}/orders/{{.Id}}"),
	api.DeleteOrderCmd:                 newCommandMapper("DELETE", "accounts/{{.Accno}}/orders/{{.Id}}"),
	api.AccountPositionsCmd:            newCommandMapper("GET", "accounts/{{.Accno}}/positions"),
	api.AccountTradesCmd:               newCommandMapper("GET", "accounts/{{.Accno}}/trades"),
	api.CountriesCmd:                   newCommandMapper("GET", "countries"),
	api.LookupCountriesCmd:             newCommandMapper("GET", "countries/{{.Countries}}"),
	api.IndicatorsCmd:                  newCommandMapper("GET", "indicators"),
	api.LookupIndicatorsCmd:            newCommandMapper("GET", "indicators/{{.Indicators}}"),
	api.SearchInstrumentsCmd:           newCommandMapper("GET", "instruments"),
	api.InstrumentsCmd:                 newCommandMapper("GET", "instruments/{{.Ids}}"),
	api.InstrumentLeveragesCmd:         newCommandMapper("GET", "instruments/{{.Id}}/leverages"),
	api.InstrumentLeverageFiltersCmd:   newCommandMapper("GET", "instruments/{{.Id}}/leverages/filters"),
	api.InstrumentOptionPairsCmd:       newCommandMapper("GET", "instruments/{{.Id}}/option_pairs"),
	api.InstrumentOptionPairFiltersCmd: newCommandMapper("GET", "instruments/{{.Id}}/option_pairs/filters"),
	api.InstrumentLookupCmd:            newCommandMapper("GET", "instruments/lookup/{{.Type}}/{{.Lookup}}"),
	api.InstrumentSectorsCmd:           newCommandMapper("GET", "instruments/sectors"),
	api.InstrumentSectorCmd:            newCommandMapper("GET", "instruments/sectors/{{.Sectors}}"),
	api.InstrumentTypesCmd:             newCommandMapper("GET", "instruments/types"),
	api.InstrumentTypeCmd:              newCommandMapper("GET", "instruments/types/{{.Type}}"),
	api.InstrumentUnderlyingsCmd:       newCommandMapper("GET", "instruments/underlyings/{{.Type}}/{.Currency}}"),
	api.ListsCmd:                       newCommandMapper("GET", "lists"),
	api.ListCmd:                        newCommandMapper("GET", "lists/{{.Id}}"),
	api.MarketsCmd:                     newCommandMapper("GET", "markets"),
	api.MarketCmd:                      newCommandMapper("GET", "markets/{{.Ids}}"),
	api.SearchNewsCmd:                  newCommandMapper("GET", "news"),
	api.NewsCmd:                        newCommandMapper("GET", "news/{{.Ids}}"),
	api.NewsSourcesCmd:                 newCommandMapper("GET", "news_sources"),
	api.RealtimeAccessCmd:              newCommandMapper("GET", "realtime_access"),
	api.TickSizesCmd:                   newCommandMapper("GET", "tick_sizes"),
	api.TickSizeCmd:                    newCommandMapper("GET", "tick_sizes/{{.Ids}}"),
	api.TradableInfoCmd:                newCommandMapper("GET", "tradables/info/{{.Ids}}"),
	api.TradableIntradayCmd:            newCommandMapper("GET", "tradables/intraday/{{.Ids}}"),
	api.TradableTradesCmd:              newCommandMapper("GET", "tradables/trades/{{.Ids}}"),
}

func NewDefaultTransport(endpoint string, user, pass, rawPem []byte) (res api.Transport, err error) {
	credGenerator, err := crypto.NewCredentialsGenerator(user, pass, rawPem)
	if err != nil {
		return
	}
	fmt.Print(credGenerator())

	res = func(req *api.Request) (res api.Response) {
		defer func() {
			if err := recover(); err != nil {
				res.Fail(-1, fmt.Sprintf("%+v", err))
			}
		}()

		if templ := CommandTemplates[req.Command]; templ != nil {
			qmap, err := req.QueryMap()
			if err != nil {
				panic(err) // TODO: fix
			}
			path, err := templ.execute(qmap)
			if err != nil {
				panic(err) // TODO: fix
			}
			fmt.Printf("Got my path: %+v\n", path)
		} else {
			panic(fmt.Errorf("Unable to find template for %s", req.Command))
		}
		return
	}
	return
}
