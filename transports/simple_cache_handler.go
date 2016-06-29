package transports

import (
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/api"
	"log"
	"sync"
	"time"
)

//type TransportCacheHandler func(RequestCommandInfo, TransportHandler, *Request) (Response)

type cachedEntry struct {
	timestamp time.Time
	data      json.RawMessage
}

type SimpleMemoryCacheHandler struct {
	sync.RWMutex
	memMap map[string]cachedEntry
}

func NewSimpleMemoryCacheHandler() *SimpleMemoryCacheHandler {
	return &SimpleMemoryCacheHandler{memMap: make(map[string]cachedEntry)}
}

// Implements TransportCacheHandler func(RequestCommandInfo, TransportHandler, *Request) (Response)
func (smch *SimpleMemoryCacheHandler) Handle(info api.RequestCommandInfo, th api.TransportHandler, req *api.Request) (res api.Response) {
	if info.TimeToLive > 0 {
		params := req.Args.SubParams(info.GetArgumentNames()...)
		params["cmd"] = string(info.Command)
		if b, err := json.Marshal(params); err != nil {
			res.Fail(-8080, fmt.Sprintf("Unable to make a key from %+v: %+v", req, err))
		} else {
			smch.RLock()
			entry, ok := smch.memMap[string(b)]
			smch.RUnlock() // Unlock now.  Dont care if we get a newer result in paralell
			if ok {
				eol := entry.timestamp.Add(time.Duration(info.TimeToLive) * time.Millisecond)
				log.Printf("Found entry with end of life at [%s] for '%s'", eol.String(), string(b))
				if eol.After(time.Now()) {
					res.Payload = entry.data
					return // Return cached entry
				} else {
					log.Printf("Entry too old, will refresh cache")
				}
			}

			res = th.Preform(req)
			if !res.IsError() {
				smch.Lock()
				defer smch.Unlock()
				smch.memMap[string(b)] = cachedEntry{time.Now(), res.Payload}
			}
		}
	} else {
		res = th.Preform(req)
	}
	return
}
