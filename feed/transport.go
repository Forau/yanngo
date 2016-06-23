// TODO: Renaming of stuff
package feed

import (
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/remote"
	"log"
	"math/rand"
	"sync"
)

type tradeState struct {
	sync.RWMutex
	state map[FeedSubscriptionKey]map[string]interface{} // Key, and the json as map[string]interface{}
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
	infoMap map[string]string // Only for info.
	dstChan remote.StreamTopicChannel
	subs    []*FeedCmd
	writer  CmdWriter

	tradeState tradeState
}

func NewFeedTransport(dstChan remote.StreamTopicChannel) *FeedState {
	return &FeedState{
		dstChan:    dstChan,
		tradeState: tradeState{state: make(map[FeedSubscriptionKey]map[string]interface{})},
		infoMap:    make(map[string]string),
	}
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
			fs.sendToTopic(msg)
		}
	case "indicator":
		fs.sendToTopic(msg)
	case "news":
		fs.sendToTopic(msg)
	case "trading_status":
		fs.sendToTopic(msg)
	case "heartbeat":
	default:
		log.Printf("Unable to handle msg of type %s: %+v", msg.Type, msg.String())
	}
}

// Implement api.TransportHandler for our admin functionality
func (fs *FeedState) Preform(req *api.Request) (res api.Response) {
	log.Printf(" --> Got admin request: %+v", req)
	params := req.Args

	switch req.Command {
	case api.TransportRespondsToCmd:
		cmds := []api.RequestCommand{api.FeedSubCmd, api.FeedUnsubCmd, api.FeedStatusCmd, api.FeedLastCmd}
		res.Success(cmds)
	case api.FeedSubCmd:
		subKey := &FeedSubscriptionKey{T: params["type"], I: params["id"], M: params["market"]}
		cmd, err := subKey.ToFeedCmd("subscribe")
		resData := make(map[string]interface{})
		if err == nil {
			resData["subId"] = fs.AddSubscription(cmd)
			err = fs.sendCommand(cmd)
		}
		if err != nil {
			res.Fail(-57, err.Error())
		} else {
			if data, ok := fs.tradeState.get(subKey); ok {
				resData["last"] = data
			}
			res.Success(resData)
		}
	case api.FeedLastCmd:
		subKey := &FeedSubscriptionKey{T: params["type"], I: params["id"], M: params["market"]}
		if data, ok := fs.tradeState.get(subKey); ok {
			res.Success(data)
		} else {
			res.Fail(-59, "Item not found")
		}
	case api.FeedUnsubCmd:
		res.Fail(-58, "Not implemented yet")
	case api.FeedStatusCmd:
		resMap := make(map[string]interface{})
		for k, v := range fs.infoMap {
			resMap[k] = v
		}
		resMap["subsctiptions"] = fs.subs
		res.Success(resMap)
	}

	return
}
