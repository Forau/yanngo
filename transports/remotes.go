package transports

import (
	"encoding/json"
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/remote"

	"bytes"
	//	"log"
)

func NewRemoteTransportClient(rrchan remote.RequestReplyChannel) (transp api.Transport) {
	transp = func(req *api.Request) (res api.Response) {
		data, err := req.Encode()
		if err != nil {
			res.Fail(-42, err.Error())
		} else {
			resData, err := rrchan(data)
			if err != nil {
				res.Fail(-43, err.Error())
			}

			dec := json.NewDecoder(bytes.NewReader(resData))
			dec.UseNumber()
			err = dec.Decode(&res)

			//			err = json.Unmarshal(resData, &res)
			if err != nil {
				res.Fail(-44, err.Error())
			}
			//			log.Printf("Response :: %+v", string(res.Payload))
		}
		return
	}
	return
}

func BindRemoteTransportServer(topic string, pubsub remote.PubSub, transp api.TransportHandler) error {
	srh := remote.MakeSubReplyHandler(pubsub, remote.SubReplyHandlerHelperFn(func(topic string, msg []byte) (rb []byte, err error) {
		var req api.Request
		dec := json.NewDecoder(bytes.NewReader(msg))
		dec.UseNumber()
		err = dec.Decode(&req)

		//		err = json.Unmarshal(msg, &req)
		if err == nil {
			res := transp.Preform(&req)
			rb, err = json.Marshal(&res)
		}
		return
	}))
	return pubsub.Sub(topic, srh)
}
