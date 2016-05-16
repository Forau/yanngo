package transports_test

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/transports"
	"testing"

	"io/ioutil"
	"os"
)

var pemData []byte

func init() {
	file, err := os.Open("../NEXTAPI_TEST_public.pem")
	if err != nil {
		panic(err)
	}
	pemData, err = ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
}

func genCmd(cmd api.RequestCommand, q interface{}) *api.Request {
	req, err := api.NewRequest(cmd, q)
	if err != nil {
		panic(err)
	}
	return req
}

func TestCommands(t *testing.T) {
	tr, err := transports.NewDefaultTransport("end", []byte("kalle"), []byte("hemlig"), pemData)
	if err != nil {
		t.Fatal(err)
	}

	cmds := []*api.Request{
		genCmd(api.SystemStatusCmd, nil),
		genCmd(api.LoginCmd, nil),
		genCmd(api.LogoutCmd, nil),
		genCmd(api.TouchCmd, nil),
		genCmd(api.AccountsCmd, nil),
		genCmd(api.AccountCmd, nil),
		genCmd(api.AccountLedgersCmd, nil),
		genCmd(api.AccountOrdersCmd, nil),
		genCmd(api.CreateOrderCmd, nil),
		genCmd(api.ActivateOrderCmd, nil),
		genCmd(api.UpdateOrderCmd, nil),
		genCmd(api.DeleteOrderCmd, map[string]int{"Accno": 123, "Id": 321}),
		genCmd(api.AccountPositionsCmd, nil),
		genCmd(api.AccountTradesCmd, nil),
		genCmd(api.CountriesCmd, nil),
		genCmd(api.LookupCountriesCmd, nil),
		genCmd(api.IndicatorsCmd, nil),
		genCmd(api.LookupIndicatorsCmd, nil),
		genCmd(api.SearchInstrumentsCmd, nil),
		genCmd(api.InstrumentsCmd, nil),
		genCmd(api.InstrumentLeveragesCmd, nil),
		genCmd(api.InstrumentLeverageFiltersCmd, nil),
		genCmd(api.InstrumentOptionPairsCmd, nil),
		genCmd(api.InstrumentOptionPairFiltersCmd, nil),
		genCmd(api.InstrumentLookupCmd, nil),
		genCmd(api.InstrumentSectorsCmd, nil),
		genCmd(api.InstrumentSectorCmd, nil),
		genCmd(api.InstrumentTypesCmd, nil),
		genCmd(api.InstrumentTypeCmd, nil),
		genCmd(api.InstrumentUnderlyingsCmd, nil),
		genCmd(api.ListsCmd, nil),
		genCmd(api.ListCmd, nil),
		genCmd(api.MarketsCmd, nil),
		genCmd(api.MarketCmd, nil),
		genCmd(api.SearchNewsCmd, nil),
		genCmd(api.NewsCmd, nil),
		genCmd(api.NewsSourcesCmd, nil),
		genCmd(api.RealtimeAccessCmd, nil),
		genCmd(api.TickSizesCmd, nil),
		genCmd(api.TickSizeCmd, nil),
		genCmd(api.TradableInfoCmd, nil),
		genCmd(api.TradableIntradayCmd, nil),
		genCmd(api.TradableTradesCmd, nil),
	}

	for _, req := range cmds {
		res := tr(req)
		t.Log(res)
	}
}
