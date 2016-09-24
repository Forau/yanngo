// TODO: Renaming of stuff
package feed

import (
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/remote"
	//	"github.com/Forau/yanngo/swagger"  // Not using swagger.Order, to save a few lines of code
	"log"
	"math/rand"
	"sync"
	"time"

	"bytes"
)

type tradeState struct {
	sync.RWMutex
	state  map[FeedSubscriptionKey]map[string]interface{}
	orders map[string]map[string]interface{}
}

func unmarshalToMap(data []byte) (ret map[string]interface{}, err error) {
	ret = make(map[string]interface{})
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err = dec.Decode(&ret)
	return
}

func (ts *tradeState) merge(msg *FeedMsg) (ret *FeedMsg, err error) {
	if payload, err2 := unmarshalToMap(msg.Data); err2 != nil {
		err = err2
	} else {
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

func (ts *tradeState) mergeOrder(msg *FeedMsg) (ret *FeedMsg, err error) {
	if order, err2 := unmarshalToMap(msg.Data); err2 != nil {
		err = err2
	} else {
		ts.Lock()
		defer ts.Unlock()
		key := fmt.Sprintf("%v:%v", order["accno"], order["order_id"])
		if current, ok := ts.orders[key]; ok {
			for k, v := range order {
				current[k] = v
			}
			ret = &FeedMsg{Type: msg.Type}
			ret.Data, err = json.Marshal(current)
		} else {
			ts.orders[key] = order
			ret = msg
		}
	}
	return
}

func (ts *tradeState) getOrders() (res []map[string]interface{}) {
	ts.RLock()
	defer ts.RUnlock()
	for _, order := range ts.orders {
		res = append(res, order)
	}
	return
}

func (ts *tradeState) getStateKeys() (res []FeedSubscriptionKey) {
	ts.RLock()
	defer ts.RUnlock()
	for st, _ := range ts.state {
		res = append(res, st)
	}
	return
}

func (ts *tradeState) get(fsk *FeedSubscriptionKey) (ret map[string]interface{}, ok bool) {
	ts.RLock()
	defer ts.RUnlock()
	ret, ok = ts.state[*fsk]
	return
}

func (ts *tradeState) Info() map[string]interface{} {
	res := make(map[string]interface{})
	res["orders"] = ts.getOrders()
	res["states"] = ts.getStateKeys()
	return res
}

// Mainly used for info. Will be used to detect 'broken' feeds too...
type heartbeatTracker struct {
	LastPrivateHb   time.Time
	LastPublicHb    time.Time
	NumPublicHbs    int64
	NumPrivateHbs   int64
	LastPrivateHbms time.Duration
	LastPublicHbms  time.Duration
}

func (hbt *heartbeatTracker) RegisterHeartbeat(ft FeedType) {
	if ft == PrivateFeedType {
		hbt.NumPrivateHbs++
		last := hbt.LastPrivateHb
		hbt.LastPrivateHb = time.Now()
		hbt.LastPrivateHbms = hbt.LastPrivateHb.Sub(last) / time.Millisecond
	} else {
		hbt.NumPublicHbs++
		last := hbt.LastPublicHb
		hbt.LastPublicHb = time.Now()
		hbt.LastPublicHbms = hbt.LastPublicHb.Sub(last) / time.Millisecond
	}
}

func (hbt *heartbeatTracker) Info() map[string]interface{} {
	res := make(map[string]interface{})
	res["LastPriv"] = time.Now().Sub(hbt.LastPrivateHb) / time.Millisecond
	res["LastPub"] = time.Now().Sub(hbt.LastPublicHb) / time.Millisecond
	res["NumPriv"] = hbt.NumPrivateHbs
	res["NumPub"] = hbt.NumPublicHbs
	res["PrivMs"] = hbt.LastPrivateHbms
	res["PubMs"] = hbt.LastPublicHbms
	return res
}

type subscriptionHolder struct {
	sub *FeedCmd
	ids []string
}

type FeedState struct {
	api.RequestCommandTransport
	infoMap               map[string]string // Only for info.
	hbt                   heartbeatTracker
	dstChan               remote.StreamTopicChannel
	subs                  []*FeedCmd
	pubWriter, privWriter CmdWriter

	tradeState tradeState
}

func NewFeedTransport(dstChan remote.StreamTopicChannel) *FeedState {
	fs := &FeedState{
		dstChan: dstChan,
		tradeState: tradeState{
			state:  make(map[FeedSubscriptionKey]map[string]interface{}),
			orders: make(map[string]map[string]interface{}),
		},
		infoMap:                 make(map[string]string),
		RequestCommandTransport: make(api.RequestCommandTransport),
	}
	fs.init()
	go fs.monitor()
	return fs
}

func (fs *FeedState) monitor() {
	// TODO, make an exit
	for {
		time.Sleep(time.Second * 10)
		if fs.privWriter != nil && fs.hbt.LastPrivateHb.Add(time.Second*10).Before(time.Now()) {
			log.Printf("Missed a private ping. Pinging our self...: %+v", fs.hbt.Info())
			fs.privWriter(&FeedCmd{Cmd: "heartbeat"})
		}

		if fs.pubWriter != nil && fs.hbt.LastPublicHb.Add(time.Second*10).Before(time.Now()) {
			log.Printf("Missed a public ping. Pinging our self...: %+v", fs.hbt.Info())
			fs.pubWriter(&FeedCmd{Cmd: "heartbeat"})
		}

	}
}

func (fs *FeedState) SetInfo(key, val string) *FeedState {
	fs.infoMap[key] = val
	return fs
}

// Implement feed.Callback
func (fs *FeedState) OnConnect(w CmdWriter, ft FeedType) {
	log.Printf("Connect[%v]: %+v", ft, w)
	fs.infoMap[fmt.Sprintf("%d:connect_%d", ft, time.Now().Unix())] = time.Now().String()
	fs.hbt.RegisterHeartbeat(ft)
	if ft == PublicFeedType {
		fs.pubWriter = w
		for _, s := range fs.subs {
			fs.sendCommand(s)
		}
	} else {
		fs.privWriter = w
	}
}
func (fs *FeedState) OnMessage(msg *FeedMsg, ft FeedType) {
	//	log.Printf("Msg[%v]: %+v", ft, msg.String())
	fs.handleAndSend(msg, ft)
}

func (fs *FeedState) OnError(err error, ft FeedType) {
	log.Printf("Err[%v]: %+v", ft, err)
	fs.infoMap[fmt.Sprintf("%d:error_%d", ft, time.Now().Unix())] = err.Error()
}

func (fs *FeedState) AddSubscription(cmd *FeedCmd) (res string) {
	res = fmt.Sprintf("%d", rand.Uint32())
	fs.subs = append(fs.subs, cmd)
	return
}

func (fs *FeedState) sendCommand(cmd *FeedCmd) error {
	if fs.pubWriter != nil {
		log.Printf("Sending: %+v", cmd)
		return fs.pubWriter(cmd)
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
		if msg2, err := fs.tradeState.mergeOrder(msg); err != nil {
			log.Printf("Error merging orders: %+v: %+v", msg, err)
			fs.sendToTopic(msg)
		} else {
			fs.sendToTopic(msg2)
		}
	case "price", "depth", "indicator":
		if msg2, err := fs.tradeState.merge(msg); err != nil {
			fs.sendToTopic(msg)
		} else {
			fs.sendToTopic(msg2)
		}
	case "trade":
		if ft == PrivateFeedType {
			msg.Type = "privtrade" // Rename, so we can easier se our trades in feed
			fs.sendToTopic(msg)
		} else {
			fs.tradeState.merge(msg) // Just to save
			fs.sendToTopic(msg)
		}
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
	fs.hbt.RegisterHeartbeat(ft) // Always register heartbeet
}

func (fs *FeedState) subscribe(params api.Params) (json.RawMessage, error) {
	subKey := &FeedSubscriptionKey{T: params["type"], I: params["id"], M: params["market"]}

	if subKey.T == "news" {
		fmt.Sscan(params["source"], &(subKey.S))
	}

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
		AddFullArgument("id", "Instrument id", []string{}, true).
		AddFullArgument("market", "Market id", []string{}, true).
		AddFullArgument("source", "News source", []string{}, true).
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
		resMap["heartbeats"] = fs.hbt.Info()
		resMap["state"] = fs.tradeState.Info()

		return json.Marshal(resMap)
	})

}
