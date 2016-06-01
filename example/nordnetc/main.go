package main

import (
	"github.com/Forau/gocop"
	"github.com/Forau/yanngo/api"
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

func toInt(str string) int64 {
	var res int64
	fmt.Sscan(str, &res)
	return res
}

func toIntArr(in string) (res []int64) {
	if in != "" {
		for _, v := range strings.Split(in, " ") {
			res = append(res, toInt(v))
		}
	} else {
		res = []int64{}
	}
	return
}

func toStrArr(in string) (res []string) {
	return strings.Split(in, " ")
}

func main() {
	flag.Parse()

	rpctransp := gorpc.NewRpcTransportClient(*srv)
	cli := api.NewApiClient(rpctransp)

	cp := gocop.NewCommandParser()
	world := cp.NewWorld()
	cp.ResultHandler = printResult

	world.AddSubCommand("/session").Handler(func(rc gocop.RunContext) (interface{}, error) {
		return cli.Session()
	})

	world.AddSubCommand("/account").Handler(func(rc gocop.RunContext) (interface{}, error) {
		if accno := rc.Get("accno"); accno != "" {
			return cli.Account(toInt(accno))
		} else {
			return cli.Accounts()
		}
	}).AddArgument("accno").Optional()

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

	log.Printf("Starting with PID %d, and parser %+v\n", os.Getpid(), cp)

	err := cp.MainLoop()

	log.Print("Exit main loop: ", err)

}
