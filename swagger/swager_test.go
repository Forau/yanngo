package swagger_test

import (
	"encoding/json"
	"github.com/Forau/yanngo/swagger"
	"testing"
)

func TestCreateEmptyObjects(t *testing.T) {
	arr := []interface{}{
		&swagger.Account{},
		&swagger.AccountInfo{},
		&swagger.ActivationCondition{},
		&swagger.Amount{},
		&swagger.CalendarDay{},
		&swagger.Country{},
		&swagger.ErrorResponse{},
		&swagger.Feed{},
		&swagger.Indicator{},
		&swagger.Instrument{},
		&swagger.InstrumentType{},
		&swagger.IntradayGraph{},
		&swagger.IntradayTick{},
		&swagger.Issuer{},
		&swagger.Ledger{},
		&swagger.LedgerInformation{},
		&swagger.LeverageFilter{},
		&swagger.List{},
		&swagger.LoggedInStatus{},
		&swagger.Login{},
		&swagger.Market{},
		&swagger.NewsItem{},
		&swagger.NewsPreview{},
		&swagger.NewsSource{},
		&swagger.OptionPairFilter{},
		&swagger.OptionPair{},
		&swagger.Order{},
		&swagger.OrderReply{},
		&swagger.OrderType{},
		&swagger.Position{},
		&swagger.PublicTrade{},
		&swagger.PublicTrades{},
		&swagger.RealtimeAccess{},
		&swagger.Sector{},
		&swagger.Status{},
		&swagger.TicksizeInterval{},
		&swagger.TicksizeTable{},
		&swagger.Tradable{},
		&swagger.TradableId{},
		&swagger.TradableInfo{},
		&swagger.Trade{},
		&swagger.UnderlyingInfo{},
		&swagger.Validity{},
	}

	for _, itm := range arr {
		data, err := json.Marshal(itm)
		if err != nil {
			t.Errorf("Unable to marshal struct %+v -> %+v", itm, err)
		} else {
			t.Log(string(data))
		}
	}
}
