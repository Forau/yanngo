package main

import (
	"github.com/Forau/gocop"
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/feed"
	"github.com/Forau/yanngo/transports/gorpc"

	"encoding/json"
	"github.com/ugorji/go/codec"

	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	srv = flag.String("srv", ":2008", "RPC server to bind to. The address daemon is bound to")

	feedApi *feed.FeedDaemon
)

func printResult(in interface{}, err error) {
	if err != nil {
		fmt.Printf("\x1b[0;31m%+v\n", err)
	} else if in != nil {

		//w := new(bytes.Buffer)
		barr := make([]byte, 0, 64)

		h := new(codec.JsonHandle)
		enc := codec.NewEncoderBytes(&barr, h)
		err := enc.Encode(in)

		//		barr, err := codec.Marshal(in)
		if err != nil {
			fmt.Printf("Could not parse to json: %+v -> %+v\n", in, err)
		} else {
			var out bytes.Buffer
			json.Indent(&out, barr, "", "  ")
			fmt.Printf("\x1b[0;35m%s\n", out.String())
			//			out.WriteTo(os.Stdout)
		}
	}
}

func toInt(str string) uint64 {
	var res uint64
	fmt.Sscan(str, &res)
	return res
}

func toFloat(str string) float64 {
	var res float64
	fmt.Sscan(str, &res)
	return res
}

func toIntArr(in string) (res []uint64) {
	if in != "" {
		for _, v := range strings.Split(in, " ") {
			res = append(res, toInt(v))
		}
	} else {
		res = []uint64{}
	}
	return
}

func toStrArr(in string) (res []string) {
	return strings.Split(in, " ")
}

// WARNING: This is _not_ intended for the client, it will be a server based operation
func createFeeds(api *api.ApiClient) (*feed.FeedDaemon, error) {
	var err error
	if feedApi == nil {
		feedApi, err = feed.NewFeedDaemon(api)
	}
	return feedApi, err
}

func main() {
	flag.Parse()

	rpctransp := gorpc.NewRpcTransportClient(*srv)
	cli := api.NewApiClient(rpctransp)

	cp := gocop.NewCommandParser()
	world := cp.NewWorld()
	cp.ResultHandler = printResult

	world.AddSubCommand("+feed").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return createFeeds(cli)
	})

	world.AddSubCommand("+sub").Handler(func(rc gocop.RunContext) (interface{}, error) {
		err := feedApi.Subscribe(rc.Get("type"), rc.Get("indicator"), rc.Get("market"))
		return fmt.Sprintf("%v", err), err
	}).AddArgument("type").AddArgument("indicator").AddArgument("market")

	world.AddSubCommand("/session").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.Session()
	})

	accBase := world.AddSubCommand("/account")

	accBase.AddSubCommand("list").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.Accounts()
	})

	accBase.AddSubCommand("info").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.Account(toInt(rc.Get("accno")))
	}).AddArgument("accno")

	accBase.AddSubCommand("ledger").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.AccountLedgers(toInt(rc.Get("accno")))
	}).AddArgument("accno")

	accBase.AddSubCommand("positions").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.AccountPositions(toInt(rc.Get("accno")))
	}).AddArgument("accno")

	accBase.AddSubCommand("trades").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.AccountTrades(toInt(rc.Get("accno")))
	}).AddArgument("accno")

	orderBase := world.AddSubCommand("/order").AddArgument("accno")

	orderBase.AddSubCommand("list").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.AccountOrders(toInt(rc.Get("accno")))
	})

	orderBase.AddSubCommand("activate").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.ActivateOrder(toInt(rc.Get("accno")), toInt(rc.Get("orderId")))
	}).AddArgument("orderId")

	orderBase.AddSubCommand("delete").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.DeleteOrder(toInt(rc.Get("accno")), toInt(rc.Get("orderId")))
	}).AddArgument("orderId")

	createOrder := orderBase.AddSubCommand("simple").Handler(func(rc gocop.RunContext) (interface{}, error) {
		var side string
		if rc.Get("buy") == "buy" {
			side = "BUY"
		} else if rc.Get("sell") == "sell" {
			side = "SELL"
		} else {
			return nil, fmt.Errorf("We did not select which side to trade")
		}
		return cli.CreateSimpleOrder(toInt(rc.Get("accno")), rc.Get("identifier"),
			toInt(rc.Get("market")), toFloat(rc.Get("price")), toInt(rc.Get("volume")), side)
	})

	createOrder.AddSubCommand("buy").AddArgument("identifier").AddArgument("market").AddArgument("price").AddArgument("volume")
	createOrder.AddSubCommand("sell").AddArgument("identifier").AddArgument("market").AddArgument("price").AddArgument("volume")

	world.AddSubCommand("/countries").Description("List countries").Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toStrArr(rc.Get("countries"))
		return cli.Countries(arr...)
	}).AddArgument("countries").Times(0, 100)

	world.AddSubCommand("/indicators").Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toStrArr(rc.Get("indicators"))
		return cli.Indicators(arr...)
	}).AddArgument("indicators").Times(0, 100)

	instruments := world.AddSubCommand("/instrument").Description("Instrument queries")
	instruments.Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toIntArr(rc.Get("instruments"))
		return cli.Instruments(arr...)
	}).AddArgument("instruments").Times(1, 100).Description("List of instrument id's")

	instruments.AddSubCommand("search").Description("Free-text search").AddArgument("query").Description("Free text").Times(1, 100).Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.InstrumentSearch(rc.Get("query"))
	})

	instruments.AddSubCommand("lookup").AddArgument("type").Optional().AddArgument("lookup").Handler(func(rc gocop.RunContext) (interface{}, error) {
		typ := rc.Get("type")
		if typ == "" {
			typ = "market_id_identifier" // Default
		}
		return cli.InstrumentLookup(typ, rc.Get("lookup"))
	})

	world.AddSubCommand("/list").Handler(func(rc gocop.RunContext) (interface{}, error) {
		list := rc.Get("list")
		if list != "" {
			return cli.List(toInt(list))
		} else {
			return cli.Lists()
		}
	}).AddArgument("list").Optional()

	world.AddSubCommand("/market").Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toIntArr(rc.Get("markets"))
		return cli.Market(arr...)
	}).AddArgument("markets").Times(0, 100)

	tradable := world.AddSubCommand("/tradable")
	tradable.Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toStrArr(rc.Get("tradables"))
		return cli.TradableInfo(arr...)
	}).AddArgument("tradables").Times(1, 100)

	tradable.AddSubCommand("day").Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toStrArr(rc.Get("tradables"))
		return cli.TradableDay(arr...)
	}).AddArgument("tradables").Times(1, 100)

	tradable.AddSubCommand("trades").Handler(func(rc gocop.RunContext) (interface{}, error) {
		arr := toStrArr(rc.Get("tradables"))
		return cli.TradableTrades(arr...)
	}).AddArgument("tradables").Times(1, 100)

	world.AddSubCommand("/realtime").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.RealtimeAccess()
	})

	world.AddSubCommand("/tick").Handler(func(rc gocop.RunContext) (interface{}, error) {
		ticks := rc.Get("ticks")
		if ticks == "" {
			return cli.TickSizes()
		} else {
			return cli.TickSize(toIntArr(ticks)...)
		}
	}).AddArgument("ticks").Times(0, 100)

	log.Printf("Starting with PID %d, and parser %+v\n", os.Getpid(), cp)

	err := cp.MainLoop()

	log.Print("Exit main loop: ", err)

}
