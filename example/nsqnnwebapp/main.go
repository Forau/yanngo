package main

import (
	"gopkg.in/igm/sockjs-go.v2/sockjs"

	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/remote"
	"github.com/Forau/yanngo/remote/nsqconn"
	"github.com/Forau/yanngo/transports"

	"encoding/json"
	"net/http"

	"flag"
	"log"
	"strings"
)

// For multiple flags
type StringArray []string

func (a *StringArray) Set(s string) error {
	*a = append(*a, s)
	return nil
}
func (a *StringArray) String() string {
	return strings.Join(*a, ",")
}

var (
	topic = flag.String("topic", "nordnet.api", "Topic server's API is listening on")
	inbox = flag.String("inbox", "", "Topic the client listens on. Use for debugging.  If not set, random is provided")
)

type RequestWithId struct {
	Id int `json:"id,omitempty"`
	api.Request
}

type ResponseWithTypeAndId struct {
	Id   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	api.Response
}

// We do not do too much error checking.  It should work...
func (rwtai *ResponseWithTypeAndId) Stringify() string {
	b, err := json.Marshal(rwtai)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func main() {
	var nsqIps StringArray
	flag.Var(&nsqIps, "nsqd", "NSQD ip's. (Can be used multiple times for each nsqd)")

	flag.Parse()

	nsqb := nsqconn.NewNsqBuilder()
	nsqb.AddNsqdIps(nsqIps...)

	log.Printf("Added IP's: %+v", nsqIps)

	nsqd, err := nsqb.Build()
	if err != nil {
		panic(err)
	}

	var pubsub remote.ReplyablePubSub
	if *inbox != "" {
		pubsub, err = remote.NewReplyablePubSubWithInbox(nsqd, *inbox)
	} else {
		pubsub, err = remote.NewReplyablePubSub(nsqd)
	}
	if err != nil {
		panic(err)
	}

	// Make transport to the server
	rchan := remote.MakeRequestReplyChannel(pubsub, *topic)
	rtrans := transports.NewRemoteTransportClient(rchan)

	handlerSockJs := sockjs.NewHandler("/sockjs", sockjs.DefaultOptions, func(session sockjs.Session) {
		log.Printf("New SockJS session: %+v", session)
		for {
			log.Printf("Waiting for msg.....: %+v", session)
			if inMsg, err := session.Recv(); err == nil {
				go func(msg string) { // Take msg as argument, so we dont close over changing value
					var req RequestWithId
					var res ResponseWithTypeAndId
					err := json.Unmarshal([]byte(msg), &req)
					if err != nil {
						res.Fail(-99, err.Error())
					} else {
						res.Type = string(req.Command)
						res.Id = req.Id
						res.Response = rtrans.Preform(&req.Request)
					}
					session.Send(res.Stringify())
				}(inMsg)
			} else {
				log.Printf("Error in SockJS, breaking loop: %+v", err)
				break
			}
		}
	})
	http.Handle("/sockjs/", handlerSockJs)

	feedSubs := []sockjs.Session{}
	pubsub.Sub("nordnet.feed", remote.SubHandlerFn(func(topic string, data []byte) error {
		log.Printf("Got feed %s", string(data))

		// Reverse iterate, so we can delete items that gave errors
		for i := len(feedSubs) - 1; i >= 0; i-- {
			feed := feedSubs[i]
			err := feed.Send(string(data))
			if err != nil {
				// Delete current element
				feedSubs = append(feedSubs[:i], feedSubs[i+1:]...)
			}
		}
		return nil
	}))

	handlerFeedJs := sockjs.NewHandler("/feed", sockjs.DefaultOptions, func(session sockjs.Session) {
		log.Printf("New feed session: %+v", session)
		feedSubs = append(feedSubs, session)
		for {
			if inMsg, err := session.Recv(); err == nil {
				log.Printf("Got message in feed, ignoring.... -> %+v", inMsg)
			} else {
				log.Printf("Error in SockJS, breaking loop: %+v", err)
				break
			}
		}
	})
	http.Handle("/feed/", handlerFeedJs)

	http.Handle("/", http.FileServer(http.Dir("web/")))

	log.Fatal(http.ListenAndServe(":8081", nil))
}
