package remote

import (
	"encoding/json"
	"fmt"
	"log"

	"math/rand"
	"sync"
	"time"
)

// Generic handler for subscriptions
type SubHandler interface {
	Handle(topic string, data []byte) error
}
type SubHandlerFn func(topic string, data []byte) error

func (sh SubHandlerFn) Handle(topic string, data []byte) error {
	return sh(topic, data)
}

// Generic PubSub interface.
type PubSub interface {
	Pub(topic string, data []byte) error
	Sub(topic string, handler SubHandler) error
	Close() error
}

type StreamTopicChannel func(data []byte) error

func MakeStreamTopicChannel(ps PubSub, topic string) StreamTopicChannel {
	return func(data []byte) error {
		return ps.Pub(topic, data)
	}
}

// Simple message with some metadata for reply
type ReplyableMessage struct {
	MsgId   int64
	ReplyTo string
	Payload []byte
}

func (rm *ReplyableMessage) Encode() ([]byte, error) {
	return json.Marshal(rm)
}

// The reply message to recieve
type MessageReply struct {
	MsgId   int64
	Payload []byte
	Error   string
}

func (mr *MessageReply) Encode() ([]byte, error) {
	return json.Marshal(mr)
}

// If we use this SubHandler function, then we will handle ReplyableMessage's, and can do basic RPC. We assume json encoding...
type SubReplyHandlerHelperFn func(topic string, msg []byte) ([]byte, error)

// A func to handle msg reply. Will only bee needed for pubsub's where it is not included.
func MakeSubReplyHandler(ps PubSub, fn SubReplyHandlerHelperFn) SubHandler {
	return SubHandlerFn(func(topic string, data []byte) (err error) {
		var msg ReplyableMessage
		err = json.Unmarshal(data, &msg)
		if err == nil {
			reply := &MessageReply{MsgId: msg.MsgId}

			b, err := fn(topic, msg.Payload)
			if err != nil {
				reply.Error = err.Error()
			} else {
				reply.Payload = b
			}

			repb, err := reply.Encode()
			if err != nil {
				// TODO: Investigate if we get this
				log.Printf("Unable to encode reply: %+v", reply)
			} else {
				err := ps.Pub(msg.ReplyTo, repb)
				if err != nil {
					log.Printf("Unable to reply on message %+v with %+v", msg, reply)
				}
			}
		}
		return
	})
}

// Bare function for request reply. Will be mapped 1 to 1 with a topic/endpoints.
type RequestReplyChannel func([]byte) ([]byte, error)

// A wrapper to make a RequestReplyChannel. It the rpc/eventbus has native request reply, then that implementation will have its own RequestReplyChannel creator.
func MakeRequestReplyChannel(rps ReplyablePubSub, topic string) RequestReplyChannel {

	return func(data []byte) ([]byte, error) {
		res, err := rps.Request(topic, data)
		log.Printf("MakeRequestReplyChannel:: %+v, %+v", res, err)
		if err != nil {
			return []byte{}, err
		} else {
			return res.Payload, nil
		}
	}
}

type ReplyablePubSub interface {
	PubSub
	Request(string, []byte) (*MessageReply, error)
}

// Will implement PubSub, and add functions for ReplyableMessage
type replyablePubSub struct {
	ps PubSub

	sync.RWMutex // For the reply map
	replies      map[int64]chan<- *MessageReply
	replyTopic   string
}

func NewReplyablePubSub(ps PubSub) (rep ReplyablePubSub, err error) {
	return NewReplyablePubSubWithInbox(ps, fmt.Sprintf("INBOX.%d", rand.Int63()))
}

func NewReplyablePubSubWithInbox(ps PubSub, inbox string) (rep ReplyablePubSub, err error) {
	repOb := &replyablePubSub{ps: ps, replies: make(map[int64]chan<- *MessageReply), replyTopic: inbox}
	err = repOb.Sub(repOb.replyTopic, repOb)
	rep = repOb
	return
}

// Implement PubSub by just forward to PubSub
func (rps *replyablePubSub) Pub(topic string, data []byte) error {
	return rps.ps.Pub(topic, data)
}

// Implement PubSub by just forward to PubSub
func (rps *replyablePubSub) Sub(topic string, handler SubHandler) error {
	return rps.ps.Sub(topic, handler)
}

// Implement PubSub by just forward to PubSub
func (rps *replyablePubSub) Close() error {
	return rps.ps.Close()
}

func (rps *replyablePubSub) Handle(topic string, data []byte) (err error) {
	// Check to see if it is a reply
	var reply MessageReply
	err = json.Unmarshal(data, &reply)
	if err != nil {
		log.Printf("[%s] Unable to handle message %+v: %+v", topic, data, err)
	} else {
		rps.remove(reply.MsgId, &reply)
	}
	return
}

func (rps *replyablePubSub) remove(mid int64, msg *MessageReply) {
	rps.RWMutex.Lock()
	defer rps.RWMutex.Unlock()
	defer func() {
		if err := recover(); err != nil {
			log.Printf("WARN: %+v", err)
		}
	}()

	ch, ok := rps.replies[mid]
	delete(rps.replies, mid)
	if ok {
		ch <- msg
		close(ch)
	}
}

func (rps *replyablePubSub) sendReplyableMessage(topic string, data []byte) (msgId int64, ch <-chan *MessageReply, err error) {
	msgId = rand.Int63()
	ch0 := make(chan *MessageReply, 1)
	ch = ch0
	msg := &ReplyableMessage{MsgId: msgId, Payload: data}
	rps.RWMutex.Lock()         // Expencive, but for now, lets lock
	defer rps.RWMutex.Unlock() // We could unlock sooner, but better safe then sorry

	rps.replies[msg.MsgId] = ch0
	msg.ReplyTo = rps.replyTopic
	b, err2 := msg.Encode()
	if err2 != nil {
		rps.remove(msg.MsgId, nil)
		return -1, nil, err2
	}
	err = rps.Pub(topic, b)
	if err != nil {
		rps.remove(msg.MsgId, nil)
	}
	return
}

// NATS have built in request, but NSQ doesnt, so lets make a simple wrapper
func (rps *replyablePubSub) Request(topic string, data []byte) (res *MessageReply, err error) {
	msgId, ch, err2 := rps.sendReplyableMessage(topic, data)
	if err2 != nil {
		return nil, err2
	}

	select {
	case res = <-ch:
		if res == nil {
			err = fmt.Errorf("Not no response to request: %d", msgId)
		}
	case <-time.After(time.Millisecond * 30000):
		err = fmt.Errorf("Timeout: No reply in 30 sec")
		rps.remove(msgId, nil)
	}
	return
}
