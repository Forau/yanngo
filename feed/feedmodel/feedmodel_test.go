package feedmodel_test

import (
	"github.com/Forau/yanngo/feed"
	"github.com/Forau/yanngo/feed/feedmodel"

	"bufio"
	"os"

	"testing"
)

func TestPriceDataSet(t *testing.T) {
	var priceDataSet *feedmodel.FeedPriceData

	fcli := feed.FeedClient(func(msg *feedmodel.FeedMsg) {
		if msg.Type == "price" {
			var price feedmodel.FeedPriceData
			if err := msg.DecodeData(&price); err == nil {
				priceDataSet = priceDataSet.Append(&price)
			} else {
				t.Errorf("Could not decode price data: %+v", err)
			}
		}
	})

	file, err := os.Open("testdata.feed")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fcli.Handle("testtopic", []byte(scanner.Text()))
	}

	pdSlice := priceDataSet.ToSlice()
	t.Logf("FeedPriceData read %d entries: %+v to %+v.", len(pdSlice), pdSlice[0], pdSlice[len(pdSlice)-1])

	if last3Trades := priceDataSet.ToLimitedSlice(feedmodel.FilterTrades().Filter(feedmodel.FilterCount(3))); len(last3Trades) == 3 {
		t.Logf("Last 3 trades: \n%+v\n%+v\n%+v", last3Trades[0], last3Trades[1], last3Trades[2])
	} else {
		t.Errorf("Expected 3 trades, but got %d trades: %+v", len(last3Trades), last3Trades)
	}

	last3IgnoreSameTickFilter := feedmodel.FilterTrades().Filter(feedmodel.FilterOnlyLastPerTickTimestamp()).Filter(feedmodel.FilterCount(3))
	if last3Trades := priceDataSet.ToLimitedSlice(last3IgnoreSameTickFilter); len(last3Trades) == 3 {
		t.Logf("Last 3 trades with different ticks: \n%+v\n%+v\n%+v", last3Trades[0], last3Trades[1], last3Trades[2])
	} else {
		t.Errorf("Expected 3 trades with different ticks, but got %d trades: %+v", len(last3Trades), last3Trades)
	}

	prices := []float64{}
	priceDataSet.Each(func(in *feedmodel.FeedPriceData) {
		prices = append(prices, in.Last)
	})
	t.Logf("Prices: %+v", prices)

	var lastTrade *feedmodel.FeedPriceData
	for _, fp := range priceDataSet.ToLimitedSlice(feedmodel.FilterTrades().Filter(feedmodel.FilterOnlyLastPerTickTimestamp())) {
		if lastTrade != nil {
			volDiff := fp.Turnover_volume - lastTrade.Turnover_volume
			t.Logf("Trade: %f %f%%. %fsec  %d", fp.Last, (1-(lastTrade.Last/fp.Last))*100, float64(fp.Trade_timestamp-lastTrade.Trade_timestamp)/1000, volDiff)
		}
		lastTrade = fp
	}

}
