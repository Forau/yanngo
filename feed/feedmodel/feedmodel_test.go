package feedmodel_test

import (
	"github.com/Forau/yanngo/feed"
	"github.com/Forau/yanngo/feed/feedmodel"
	"github.com/Forau/yanngo/remote"

	"bufio"
	"os"
	//	"time"
	"fmt"

	"testing"
)

func streamFileToSubHandler(fname string, sh remote.SubHandler) error {
	if file, err := os.Open(fname); err != nil {
		return err
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			sh.Handle("testtopic", []byte(scanner.Text()))
		}
	}
	return nil
}

func TestPriceDataSet(t *testing.T) {
	var priceDataSet *feedmodel.FeedPriceData
	var tradeDataSet []*feedmodel.FeedTradeData
	fcli := feed.FeedClient(func(msg *feedmodel.FeedMsg) {
		if msg.Type == "price" {
			var price feedmodel.FeedPriceData
			if err := msg.DecodeData(&price); err == nil {
				priceDataSet = priceDataSet.Append(&price)
			} else {
				t.Errorf("Could not decode price data: %+v", err)
			}
		} else if msg.Type == "trade" {
			var trade feedmodel.FeedTradeData
			if err := msg.DecodeData(&trade); err == nil {
				tradeDataSet = append(tradeDataSet, &trade)
			} else {
				t.Errorf("Could not decode trade data: %+v", err)
			}
		}
	})

	if err := streamFileToSubHandler("testdata.feed", fcli); err != nil {
		t.Fatal(err)
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
			t.Logf("Trade: %f %f%%. %fsec  %d  %v", fp.Last, (1-(lastTrade.Last/fp.Last))*100, float64(fp.Trade_timestamp-lastTrade.Trade_timestamp)/1000, volDiff, fp.TradeSide())
		}
		lastTrade = fp
	}

	t.Logf("Trades found in feed: %d\n", len(tradeDataSet))
	t.Logf("Trades found in price filter: %d\n", len(priceDataSet.ToLimitedSlice(feedmodel.FilterTrades())))

}

// For now we dont test price and volume.
func TestStreamTrades(t *testing.T) {
	trades := make(map[string][]feedmodel.FeedTradeData)

	callback := func(trade *feedmodel.FeedTradeData) {
		t.Logf("Got %v\n", trade)
		key := fmt.Sprintf("%s:%d-%d-%d", trade.Identifier, trade.Market, trade.Volume, trade.Trade_timestamp)
		if arr, ok := trades[key]; ok {
			trades[key] = append(arr, *trade)
		} else {
			trades[key] = []feedmodel.FeedTradeData{*trade}
		}
	}
	fcli := feedmodel.NewTradeCatcher(callback)

	fcliSorted := feed.FeedSorterDelayedIfNotSeqId(5000, fcli.OnFeed)

	if err := streamFileToSubHandler("testdata.feed", feed.FeedClient(fcliSorted)); err != nil {
		t.Fatal(err)
	}

	count := 0
	for k, v := range trades {
		count++
		if len(v)%2 != 0 || len(v) == 0 {
			t.Errorf("Expected even number of trades for %s, but got %d: %v", k, len(v), v)
		}
	}
	t.Logf("Got %d trades", count)
}

func TestStreamIndicators(t *testing.T) {
	var indDax, indOmx feedmodel.FeedIndicatorDataSequence
	fcli := func(msg *feedmodel.FeedMsg) {
		var ind feedmodel.FeedIndicatorData
		if err := msg.DecodeData(&ind); err != nil {
			t.Fatal(err)
		} else {
			switch ind.Identifier {
			case "DAX":
				indDax = append(indDax, ind)
			case "OMXS30":
				indOmx = append(indOmx, ind)
			}
		}
	}

	fcliSorted := feed.FeedSorterDelayedIfNotSeqId(5000, fcli)

	if err := streamFileToSubHandler("testdata.feed.ind", feed.FeedClient(fcliSorted)); err != nil {
		t.Fatal(err)
	}
	t.Logf("All done: %d - %d\n", len(indDax), len(indOmx))
	indDax.Sort()
	t.Logf("LastDax: %+v", indDax.Last())
	t.Logf("At DAX 1483009130777: %+v\n", indDax.At(1483009130777))
	t.Logf("At DAX 1000000000000: %+v\n", indDax.At(1000000000000))
	t.Logf("LastOmx: %+v", indOmx.Last())
}
