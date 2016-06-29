// TODO: Renaming of stuff
package feed

import (
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/remote"
	"github.com/Forau/yanngo/swagger"
	"log"
	"math/rand"
	"sync"
)

type tradeState struct {
	sync.RWMutex
	state  map[FeedSubscriptionKey]map[string]interface{}
	orders map[int64]swagger.Order
}

func (ts *tradeState) merge(msg *FeedMsg) (ret *FeedMsg, err error) {
	var payload map[string]interface{}
	err = json.Unmarshal(msg.Data, &payload)
	if err == nil {
		key := FeedSubscriptionKey{T: msg.Type, I: fmt.Sprintf("%v", payload["i"]), M: fmt.Sprintf("%v", payload["m"])}
		ts.Lock()
		defer ts.Unlock()
		if current, ok := ts.state[key]; ok {
			for k, v := range payload {
				current[k] = v
			}
			ret = &FeedMsg{Type: msg.Type}
			ret.Data, err = json.Marshal(current)
		} else {
			ts.state[key] = payload
			ret = msg
		}
	}
	return
}

func (ts *tradeState) mergeOrder(msg *FeedMsg) (err error) {
	var order swagger.Order
	err = json.Unmarshal(msg.Data, &order)
	if err == nil {
		ts.Lock()
		defer ts.Unlock()
		ts.orders[order.OrderId] = order
	}
	return
}

func (ts *tradeState) getOrders() (res []swagger.Order) {
	ts.RLock()
	defer ts.RUnlock()
	for _, order := range ts.orders {
		res = append(res, order)
	}
	return
}

func (ts *tradeState) get(fsk *FeedSubscriptionKey) (ret map[string]interface{}, ok bool) {
	ts.RLock()
	defer ts.RUnlock()
	ret, ok = ts.state[*fsk]
	return
}

type subscriptionHolder struct {
	sub *FeedCmd
	ids []string
}

type FeedState struct {
	api.RequestCommandTransport
	infoMap map[string]string // Only for info.
	dstChan remote.StreamTopicChannel
	subs    []*FeedCmd
	writer  CmdWriter

	tradeState tradeState
}

func NewFeedTransport(dstChan remote.StreamTopicChannel) *FeedState {
	fs := &FeedState{
		dstChan: dstChan,
		tradeState: tradeState{
			state:  make(map[FeedSubscriptionKey]map[string]interface{}),
			orders: make(map[int64]swagger.Order),
		},
		infoMap:                 make(map[string]string),
		RequestCommandTransport: make(api.RequestCommandTransport),
	}
	fs.init()
	return fs
}

func (fs *FeedState) SetInfo(key, val string) *FeedState {
	fs.infoMap[key] = val
	return fs
}

// Implement feed.Callback
func (fs *FeedState) OnConnect(w CmdWriter, ft FeedType) {
	log.Printf("Connect[%v]: %+v", ft, w)
	if ft == PublicFeedType {
		fs.writer = w
		for _, s := range fs.subs {
			fs.sendCommand(s)
		}
	}
}
func (fs *FeedState) OnMessage(msg *FeedMsg, ft FeedType) {
	//	log.Printf("Msg[%v]: %+v", ft, msg.String())
	fs.handleAndSend(msg, ft)
}

func (fs *FeedState) OnError(err error, ft FeedType) { log.Printf("Err[%v]: %+v", ft, err) }

func (fs *FeedState) AddSubscription(cmd *FeedCmd) (res string) {
	res = fmt.Sprintf("%d", rand.Uint32())
	fs.subs = append(fs.subs, cmd)
	return
}

func (fs *FeedState) sendCommand(cmd *FeedCmd) error {
	if fs.writer != nil {
		log.Printf("Sending: %+v", cmd)
		return fs.writer(cmd)
	} else {
		log.Printf("Unable to send command '%+v', writer not ready", cmd)
		return fmt.Errorf("Unable to send command")
	}
}

func (fs *FeedState) sendToTopic(msg interface{}) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling %+v: %+v", msg, err)
	} else {
		err := fs.dstChan(b)
		log.Printf("Pub to Msg(%s) : %+v", string(b), err)
	}
}

func (fs *FeedState) handleAndSend(msg *FeedMsg, ft FeedType) {
	switch msg.Type {
	case "order":
		if err := fs.tradeState.mergeOrder(msg); err != nil {
			log.Printf("Error merging orders: %+v: %+v", msg, err)
		}
		fs.sendToTopic(msg)
	case "price", "depth":
		if msg2, err := fs.tradeState.merge(msg); err != nil {
			fs.sendToTopic(msg)
		} else {
			fs.sendToTopic(msg2)
		}
	case "trade":
		if ft == PrivateFeedType {
			fs.sendToTopic(msg)
		} else {
			fs.tradeState.merge(msg) // Just to save
			fs.sendToTopic(msg)
		}
	case "indicator":
		fs.tradeState.merge(msg) // Just to save
		fs.sendToTopic(msg)
	case "news":
		fs.tradeState.merge(msg) // Just to save
		fs.sendToTopic(msg)
	case "trading_status":
		fs.tradeState.merge(msg) // Just to save
		fs.sendToTopic(msg)
	case "heartbeat":
	default:
		log.Printf("Unable to handle msg of type %s: %+v", msg.Type, msg.String())
	}
}

func (fs *FeedState) subscribe(params api.Params) (json.RawMessage, error) {
	subKey := &FeedSubscriptionKey{T: params["type"], I: params["id"], M: params["market"]}
	cmd, err := subKey.ToFeedCmd("subscribe")
	resData := make(map[string]interface{})
	if err == nil {
		resData["subId"] = fs.AddSubscription(cmd)
		err = fs.sendCommand(cmd)
	}
	if err != nil {
		return nil, err
	} else {
		if data, ok := fs.tradeState.get(subKey); ok {
			resData["last"] = data
		}
		return json.Marshal(resData)
	}
}
func (fs *FeedState) lastMsg(params api.Params) (json.RawMessage, error) {
	subKey := &FeedSubscriptionKey{T: params["type"], I: params["id"], M: params["market"]}
	if data, ok := fs.tradeState.get(subKey); ok {
		return json.Marshal(data)
	} else {
		return nil, fmt.Errorf("Item %+v not found", subKey)
	}
}

func (fs *FeedState) init() {
	fs.AddCommand(string(api.FeedSubCmd)).Description("Subscribe to a feed").
		AddFullArgument("type", "Type to subscribe to",
		[]string{"price", "depth", "trade", "trading_status", "indicator", "news"}, false).
		AddFullArgument("id", "Instrument id", []string{}, false).
		AddFullArgument("market", "Market id", []string{}, false).
		Handler(fs.subscribe)

	fs.AddCommand(string(api.FeedLastCmd)).Description("Last message from feed").
		AddFullArgument("type", "Type to get message from",
		[]string{"price", "depth", "trade", "trading_status", "indicator", "news"}, false).
		AddFullArgument("id", "Instrument id", []string{}, false).
		AddFullArgument("market", "Market id", []string{}, false).
		Handler(fs.lastMsg)

	fs.AddCommand("FeedGetOrders").Description("Get the cached orders from feed").
		Handler(func(params api.Params) (json.RawMessage, error) {
		orders := fs.tradeState.getOrders()
		return json.Marshal(orders)
	})

	fs.AddCommand("FeedGetState").Description("Get the cached state from feed subscriptions").
		AddFullArgument("id", "Instrument id filter", []string{}, true).
		AddFullArgument("market", "Market id filter", []string{}, true).
		Handler(func(params api.Params) (json.RawMessage, error) {
		res := [](map[string]interface{}){}
		id := params["id"]
		market := params["market"]
		for key, state := range fs.tradeState.state {
			if (id == "" || id == key.I) && (market == "" || market == key.M) {
				res = append(res, map[string]interface{}{"type": key.T, "data": state})
			}
		}
		return json.Marshal(res)
	})

	fs.AddCommand(string(api.FeedStatusCmd)).Description("Get some status").
		Handler(func(params api.Params) (json.RawMessage, error) {
		resMap := make(map[string]interface{})
		for k, v := range fs.infoMap {
			resMap[k] = v
		}
		resMap["subsctiptions"] = fs.subs
		return json.Marshal(resMap)
	})

}
