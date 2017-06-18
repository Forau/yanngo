package remote

import (
	"encoding/json"
	"fmt"
	"log"

	"math/rand"
	"sync"
	"time"

	"bytes"
)

var rnd *rand.Rand

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(src)
}

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
	Seq     int64
	NumSeq  int64
	Payload []byte
	Error   string
}

func MakeReplies(msgId int64, payload []byte) (ret []MessageReply) {
	fmt.Printf("Make reply for %d: %d\n", msgId, len(payload))
	maxSize := int64(500000)
	payloads := [][]byte{}
	l := int64(len(payload))
	for i := int64(0); i < l; i += maxSize {
		end := i + maxSize
		if end > l {
			end = l
		}
		payloads = append(payloads, payload[i:end])
	}
	fmt.Printf("Split to %d messages\n", len(payloads))

	l = int64(len(payloads))
	for idx, pl := range payloads {
		ret = append(ret, MessageReply{MsgId: msgId, Payload: pl, Seq: int64(idx + 1), NumSeq: l})
	}
	return
}

func (mr *MessageReply) Encode() ([]byte, error) {
	return json.Marshal(mr)
}

func (mr *MessageReply) String() string {
	b, err := json.Marshal(mr)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(b)
}

func (mr *MessageReply) Info() string {
	if mr == nil {
		return "MessageReply{NIL}"
	}
	return fmt.Sprintf("MessageReply{%d, Seq: %d/%d. Len %d: %s}\n", mr.MsgId, mr.Seq, mr.NumSeq, len(mr.Payload), mr.Error)
}

// If we use this SubHandler function, then we will handle ReplyableMessage's, and can do basic RPC. We assume json encoding...
type SubReplyHandlerHelperFn func(topic string, msg []byte) ([]byte, error)

// A func to handle msg reply. Will only bee needed for pubsub's where it is not included.
func MakeSubReplyHandler(ps PubSub, fn SubReplyHandlerHelperFn) SubHandler {
	return SubHandlerFn(func(topic string, data []byte) (err error) {
		var msg ReplyableMessage
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.UseNumber()
		err = dec.Decode(&msg)
		//		err = json.Unmarshal(data, &msg)

		if err == nil {
			var replies []MessageReply
			b, err := fn(topic, msg.Payload)
			if err != nil {
				replies = append(replies, MessageReply{MsgId: msg.MsgId, Error: err.Error()})
			} else {
				replies = MakeReplies(msg.MsgId, b)
			}

			for _, reply := range replies {
				fmt.Printf("Sending %v\n", reply.Info())
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
		//		log.Printf("MakeRequestReplyChannel:: %+v, %+v", res.String(), err)
		if err != nil {
			return []byte{}, err
		} else {
			return res.Payload, nil
		}
	}
}

type ReplySegmentChannel struct {
	Parts   []MessageReply
	Channel chan<- *MessageReply
}

func (rsc *ReplySegmentChannel) SendIfComplete(msg *MessageReply) (ok bool) {
	if msg == nil || msg.NumSeq == 1 {
		ok = true
		fmt.Printf("Sending 'complete' for msg %+v\n", msg.Info())
	} else {
		rsc.Parts = append(rsc.Parts, *msg)
		if newMsg := rsc.assembleParts(); newMsg != nil {
			msg = newMsg
			ok = true
			fmt.Printf("All %d parts assembled. Sending complete on %+v\n", len(rsc.Parts), msg.Info())
		} else {
			fmt.Printf("All parts not yet assembled... Current cache size: %d\n", len(rsc.Parts))
		}
	}
	if ok {
		go func() {
			rsc.Channel <- msg
			close(rsc.Channel)
		}()
	}
	return
}

func (rsc *ReplySegmentChannel) assembleParts() (msg *MessageReply) {
	if l := int64(len(rsc.Parts)); l == 0 || rsc.Parts[0].NumSeq > l {
		return nil // Early break for empty, or too few.
	} else {
		parts := rsc.Parts[0].NumSeq
		data := []byte{}
		segCount := int64(0)
		for seg := int64(1); seg <= parts; seg++ {
			for _, p := range rsc.Parts {
				if p.Seq == seg {
					data = append(data, p.Payload...)
					fmt.Printf("Appending data %d for seq %d. Total size %d\n", len(p.Payload), seg, len(data))
					segCount++
				}
			}
		}
		if segCount == parts {
			msg = &MessageReply{MsgId: rsc.Parts[0].MsgId, Seq: 1, NumSeq: 1, Payload: data}
		}
	}
	return
}

type ReplyablePubSub interface {
	PubSub
	Request(string, []byte) (*MessageReply, error)
}

// Will implement PubSub, and add functions for ReplyableMessage
type replyablePubSub struct {
	ps PubSub

	sync.RWMutex // For the reply map
	replies      map[int64]*ReplySegmentChannel
	replyTopic   string
}

func NewReplyablePubSub(ps PubSub) (rep ReplyablePubSub, err error) {
	return NewReplyablePubSubWithInbox(ps, fmt.Sprintf("INBOX.%d", rnd.Int63()))
}

func NewReplyablePubSubWithInbox(ps PubSub, inbox string) (rep ReplyablePubSub, err error) {
	repOb := &replyablePubSub{ps: ps, replies: make(map[int64]*ReplySegmentChannel), replyTopic: inbox}
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
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err = dec.Decode(&reply)
	//	err = json.Unmarshal(data, &reply)
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

	if rsc, ok := rps.replies[mid]; ok {
		if sent := rsc.SendIfComplete(msg); sent {
			delete(rps.replies, mid)
		}
	} else {
		log.Printf("Not found any listeners for message %+v\n", mid)
	}
}

func (rps *replyablePubSub) sendReplyableMessage(topic string, data []byte) (msgId int64, ch <-chan *MessageReply, err error) {
	msgId = rnd.Int63()
	ch0 := make(chan *MessageReply, 5)
	ch = ch0
	msg := &ReplyableMessage{MsgId: msgId, Payload: data}
	rps.RWMutex.Lock()         // Expencive, but for now, lets lock
	defer rps.RWMutex.Unlock() // We could unlock sooner, but better safe then sorry

	rps.replies[msg.MsgId] = &ReplySegmentChannel{Channel: ch0}
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
