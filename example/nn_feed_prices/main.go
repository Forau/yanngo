// nsq_tail -topic nordnet.feed --nsqd-tcp-address 127.0.0.1:5150 | go run main.go

package main

import (
	//  tm "github.com/buger/goterm"  //Is a nicer choise, but dont work when piped to
	"fmt"
	"os"
	"time"

	"encoding/json"
	"sync"

	"sort"
	"text/template"
)

var (
	prices        = make(map[string]map[string]interface{})
	lock          sync.RWMutex
	header        = `instrument ask   ask_volume   bid bid_volume `
	priceTemplate = `{{.m}}:{{.i | pad8 }} {{.ask | padf8}}{{.ask_volume | padi8}} {{.bid | padf8}} {{.bid_volume | padi8 }} | OHLC {{.open}} {{.high}} {{.low}} {{.close}} | Trade: p {{.last}} v {{.last_volume}} t {{.trade_timestamp| date }} | {{.tick_timestamp | date}}`
)

type feedMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func parseDate(in float64) string {
	return time.Unix(int64(in)/1000, 0).Format("2/1 15:04:05")
}

func pad8(in string) string {
	data := []byte(in)
	data = append(data, ' ', ' ', ' ', ' ', ' ', ' ', ' ')
	return string(data[:7])
}

func padi8(in interface{}) string {
	return pad8(fmt.Sprintf("%.0f", in))
}

func padf8(in interface{}) string {
	return pad8(fmt.Sprintf("%.2f", in))
}

func main() {
	priceTmpl, err := template.New("price").Funcs(template.FuncMap{
		"date":  parseDate,
		"pad8":  pad8,
		"padi8": padi8,
		"padf8": padf8,
	}).Parse(priceTemplate)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			fmt.Print("\033[H\033[2J") // Clear screen. If you use windows
			fmt.Println("Current Time:", time.Now().Format(time.RFC1123))
			fmt.Println("----------------------------------------------------------")
			lock.Lock()

			var keys []string
			for k := range prices {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			fmt.Println(header)
			for _, k := range keys {
				priceTmpl.Execute(os.Stdout, prices[k])
				fmt.Println()
			}

			lock.Unlock()
			time.Sleep(time.Second)
		}
	}()

	input := json.NewDecoder(os.Stdin)
	for {
		var msg feedMsg
		err := input.Decode(&msg)
		if err == nil && msg.Type == "price" {
			pmap := make(map[string]interface{})
			err := json.Unmarshal(msg.Data, &pmap)
			if err == nil {
				key := fmt.Sprintf("%v:%v", pmap["m"], pmap["i"])
				lock.Lock()
				prices[key] = pmap
				lock.Unlock()
			}
		}
	}
}
