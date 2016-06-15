// This package contains a simple implementation of feed that streams it to NSQD. (http://nsq.io/)
package nsqfeeder

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/feed"
	nsq "github.com/nsqio/go-nsq"

	"encoding/json"

	"fmt"
	"log"
	"os"
)

type NsqFeeder struct {
	feedTopic, adminTopic string

	producers   []*nsq.Producer // List of nsqd's to send messages to
	sendcounter int             // Just a counter to rotate used producer

	adminConsumer *nsq.Consumer

	daemon *feed.FeedDaemon

	writer feed.CmdWriter
}

func NewNsqfeeder(topic, admTop string, nsqdTCPAddrs []string, nsqConfig map[string]interface{}, api *api.ApiClient) (nf *NsqFeeder, err error) {
	nf = &NsqFeeder{
		feedTopic:  topic,
		adminTopic: admTop,
		writer:     func(cmd *feed.FeedCmd) error { log.Printf("Not connected yet"); return nil },
	}

	config := nsq.NewConfig()
	for k, v := range nsqConfig {
		config.Set(k, v)
	}

	config.UserAgent = "YANNGO - feeder v0.0"
	channel := fmt.Sprintf("yanngo_feeder_%d", os.Getpid())
	nf.adminConsumer, err = nsq.NewConsumer(nf.adminTopic, channel, config)
	if err != nil {
		return
	}

	nf.adminConsumer.AddConcurrentHandlers(nf, len(nsqdTCPAddrs))
	err = nf.adminConsumer.ConnectToNSQDs(nsqdTCPAddrs)
	if err != nil {
		return
	}

	for _, addr := range nsqdTCPAddrs {
		producer, err := nsq.NewProducer(addr, config)
		if err != nil {
			return nil, err
		}
		nf.producers = append(nf.producers, producer)
	}

	nf.daemon, err = feed.NewFeedDaemon(feed.MakePrivateSessionProvider(api), feed.MakePublicSessionProvider(api), nf)

	return
}

func (nf *NsqFeeder) HandleMessage(m *nsq.Message) error {
	// TODO
	log.Printf("Recived nsq: %+v", m)

	fcmd := &struct {
		T string `json:"t"`
		I string `json:"i"`
		M string `json:"m"`
	}{}
	err := json.Unmarshal(m.Body, fcmd)
	if err != nil {
		log.Printf("Error: %+v")
	} else {
		//    err = nf.writer(fcmd)
		nf.daemon.Subscribe(fcmd.T, fcmd.I, fcmd.M)
	}

	m.Finish()
	return err
}

func (nf *NsqFeeder) OnConnect(w feed.CmdWriter, ft feed.FeedType) {
	log.Printf("Connect[%v]: %+v", ft, w)
	if ft == feed.PublicFeedType {
		nf.writer = w
		// Send our subscriptions....
	}
}
func (nf *NsqFeeder) OnMessage(msg *feed.FeedMsg, ft feed.FeedType) {
	log.Printf("Got msg[%d]: %+v", msg, ft)
	b, err := json.Marshal(msg)
	if err != nil {
		nf.OnError(err, ft)
	} else {
		nf.sendMessage(b)
	}
}
func (nf *NsqFeeder) OnError(err error, ft feed.FeedType) {
	nf.sendMessage([]byte(fmt.Sprintf(`{"type": "error", "data": "%s"}`, err.Error())))
}

func (nf *NsqFeeder) sendMessage(msg []byte) {
	nf.sendcounter++
	go func(idx int) {
		for i := 0; i < len(nf.producers); i++ {
			mod := (i + idx) % len(nf.producers)
			prod := nf.producers[mod]
			err := prod.Publish(nf.feedTopic, msg)
			if err == nil {
				return
			} else {
				log.Printf("Error: %+v", err)
			}
		}
	}(nf.sendcounter)
}
