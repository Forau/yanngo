// A quick and dirty implementation to cache in mongodb
package mongocache

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Forau/yanngo/api"
	"log"

	"time"
)

type MongoCacheHandler struct {
	mgocol *mgo.Collection
}

func NewMongoCacheHandler(col *mgo.Collection) (mch *MongoCacheHandler, err error) {
	mch = &MongoCacheHandler{mgocol: col}
	index := mgo.Index{
		Key: []string{"cmd", "timestamp"},
	}
	err = col.EnsureIndex(index)
	return
}

// Implements TransportCacheHandler func(RequestCommandInfo, TransportHandler, *Request) (Response)
func (mch *MongoCacheHandler) Handle(info api.RequestCommandInfo, th api.TransportHandler, req *api.Request) (res api.Response) {
	if info.TimeToLive > 0 {
		freshTime := time.Now().Add(-time.Millisecond * time.Duration(info.TimeToLive))
		params := req.Args.SubParams(info.GetArgumentNames()...)
		query := bson.M{"cmd": string(info.Command), "timestamp": bson.M{"$gt": freshTime}}

		for k, v := range params {
			query[k] = v
		}

		tmpres := make(map[string]interface{})
		if err := mch.mgocol.Find(query).One(&tmpres); err == nil {
			log.Printf("got result: %v: %v -> %v", tmpres["_id"], tmpres["cmd"], tmpres["timestamp"])
			if err := res.Marshal(tmpres["payload"]); err != nil {
				log.Printf("ERROR converting payload[%+v]: %T %+v", err, tmpres["payload"], tmpres["payload"])
				res = th.Preform(req)
			}
		} else {
			log.Printf("got error: %+v", err)
			res = th.Preform(req)
			if !res.IsError() {
				query["timestamp"] = time.Now()

				// Lets unpack the payload. This makes it searchable in mongodb, and easier to inspect the cache.
				var payload interface{}
				err = res.Unmarshal(&payload)
				query["payload"] = payload

				err := mch.mgocol.Insert(query)
				if err != nil {
					log.Printf("Error insering cached entry: %+v", err)
				}
			}
		}
	} else {
		res = th.Preform(req)
	}
	return
}
