package tinypubsub

import (
	"github.com/Forau/yanngo/remote"
	"sync"
)

type TinyPubSub struct {
	sync.RWMutex
	subscriptions map[string][]remote.SubHandler
}

func NewTinyPubSub() *TinyPubSub {
	return &TinyPubSub{subscriptions: make(map[string][]remote.SubHandler)}
}

func (tps *TinyPubSub) Pub(topic string, data []byte) error {
	tps.RLock()
	defer tps.RUnlock()
	if handlers, ok := tps.subscriptions[topic]; ok {
		for _, h := range handlers {
			go h.Handle(topic, data) // We dont want to wait for it....
		}
	}
	return nil
}

func (tps *TinyPubSub) Sub(topic string, handler remote.SubHandler) error {
	tps.Lock()
	defer tps.Unlock()
	if arr, ok := tps.subscriptions[topic]; ok {
		tps.subscriptions[topic] = append(arr, handler)
	} else {
		tps.subscriptions[topic] = []remote.SubHandler{handler}
	}
	return nil
}

func (tps *TinyPubSub) Close() error {
	tps.Lock()
	defer tps.Unlock()
	tps.subscriptions = nil
	return nil
}
