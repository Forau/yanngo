package nsqconn

import (
	nsq "github.com/nsqio/go-nsq"

	"github.com/Forau/yanngo/remote"

	"fmt"
	"log"
	"math/rand"
)

type NsqConnBuilder struct {
	cfg    *nsq.Config
	errors []error

	nsqdIps    []string
	lookupdIps []string
	channel    string
}

func NewNsqBuilder() *NsqConnBuilder {
	return &NsqConnBuilder{cfg: nsq.NewConfig(), channel: fmt.Sprintf("chan%d", rand.Int63())}
}

func (ncb *NsqConnBuilder) Set(key string, val interface{}) *NsqConnBuilder {
	if err := ncb.cfg.Set(key, val); err != nil {
		ncb.errors = append(ncb.errors, err)
	}
	return ncb
}

func (ncb *NsqConnBuilder) AddNsqdIps(ips ...string) *NsqConnBuilder {
	if len(ncb.lookupdIps) > 0 {
		ncb.errors = append(ncb.errors, fmt.Errorf("Already have lookupd ips, we can not add nsqd ips as well. One or another."))
	}
	ncb.nsqdIps = append(ncb.nsqdIps, ips...)
	return ncb
}

func (ncb *NsqConnBuilder) AddLookupdIps(ips ...string) *NsqConnBuilder {
	if len(ncb.nsqdIps) > 0 {
		ncb.errors = append(ncb.errors, fmt.Errorf("Already have nsqd ips, we can not add lookupd ips as well. One or another. "))
	}
	ncb.lookupdIps = append(ncb.lookupdIps, ips...)
	return ncb
}

func (ncb *NsqConnBuilder) Channel(c string) *NsqConnBuilder {
	ncb.channel = c
	return ncb
}

func (ncb *NsqConnBuilder) Build() (remote.PubSub, error) {
	if len(ncb.errors) > 0 {
		return nil, fmt.Errorf("Errors during building nsqConn: %+v", ncb.errors)
	}

	var prod []*nsq.Producer
	// TODO: Is only nsqd allowed for producers?  Investigate
	for _, addr := range ncb.nsqdIps {
		p, err := nsq.NewProducer(addr, ncb.cfg)
		if err != nil {
			return nil, err
		} else {
			prod = append(prod, p)
		}
	}
	if len(prod) == 0 {
		return nil, fmt.Errorf("Need to add nsqd addresses so we can have producers")
	}
	return &NsqConn{builder: ncb, producers: prod}, nil
}

type NsqConn struct {
	builder *NsqConnBuilder

	consumers []*nsq.Consumer
	producers []*nsq.Producer
}

// Sumple pub
func (nc *NsqConn) Pub(topic string, data []byte) error {
	go func(idx []int) {
		for _, i := range idx {
			err := nc.producers[i].Publish(topic, data)
			if err == nil {
				return // All well
			}
		}
		log.Print("Trying to send data to '%s', but all %d produces failed", topic, len(nc.producers))
	}(rand.Perm(len(nc.producers)))
	return nil // No errors for now....
}

func (nc *NsqConn) Sub(topic string, handler remote.SubHandler) error {
	consumer, err := nsq.NewConsumer(topic, nc.builder.channel, nc.builder.cfg)
	if err != nil {
		return err
	}
	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		return handler.Handle(topic, m.Body)
	}))

	err = consumer.ConnectToNSQDs(nc.builder.nsqdIps)
	if err != nil {
		return err
	}

	err = consumer.ConnectToNSQLookupds(nc.builder.lookupdIps)
	if err != nil {
		return err
	}

	nc.consumers = append(nc.consumers, consumer)
	return nil
}

func (nc *NsqConn) Close() error {
	for _, c := range nc.consumers {
		c.Stop()
	}
	for _, p := range nc.producers {
		p.Stop()
	}
	return nil
}
