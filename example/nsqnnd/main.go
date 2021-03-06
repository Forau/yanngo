package main

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/feed"
	"github.com/Forau/yanngo/feed/feedmodel"
	"github.com/Forau/yanngo/remote"
	"github.com/Forau/yanngo/remote/nsqconn"
	"github.com/Forau/yanngo/transports"

	"io/ioutil"
	"os"

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
	user      = flag.String("user", "", "User name")
	pass      = flag.String("pass", "", "Password")
	endpoint  = flag.String("url", "https://api.test.nordnet.se/next/2", "The base URL.")
	topic     = flag.String("topic", "nordnet.api", "Topic to listen on")
	feedTopic = flag.String("feedtop", "nordnet.feed", "Topic to send feed on")
	pemFile   = flag.String("pem", "../../NEXTAPI_TEST_public.pem", "The PEM file")
)

func main() {
	var nsqIps StringArray
	flag.Var(&nsqIps, "nsqd", "NSQD ip's. (Can be used multiple times for each nsqd)")

	flag.Parse()

	file, err := os.Open(*pemFile)
	if err != nil {
		panic(err)
	}
	pem, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	baseNordnetTransport, err := transports.NewDefaultTransport(*endpoint, []byte(*user), []byte(*pass), pem)
	if err != nil {
		panic(err)
	}
	cacheHandler := transports.NewSimpleMemoryCacheHandler()
	nordnetTransport, _ := api.NewCachedTransportRouter(cacheHandler, baseNordnetTransport)

	nsqb := nsqconn.NewNsqBuilder()
	nsqb.AddNsqdIps(nsqIps...)

	log.Printf("Added IP's: %+v", nsqIps)

	nsqd, err := nsqb.Build()
	if err != nil {
		panic(err)
	}

	// TODO: Optional for server?
	pubsub, err := remote.NewReplyablePubSubWithInbox(nsqd, "INBOX.nsqnnd.client")
	if err != nil {
		panic(err)
	}

	log.Printf("We have our nsq: %+v", pubsub)

	log.Printf("We have our transport: %+v", nordnetTransport)

	err = transports.BindRemoteTransportServer(*topic, pubsub, nordnetTransport)
	if err != nil {
		panic(err)
	}

	apiCli := api.NewApiClient(nordnetTransport)
	// Feed
	feedTopicStream := remote.MakeStreamTopicChannel(nsqd, *feedTopic)
	feedCb := feed.NewFeedTransport(feedTopicStream).SetInfo("topic", *feedTopic)
	err = nordnetTransport.AddTransportHandler(feedCb)
	if err != nil {
		log.Printf("Unable to route %+v: %+v", feedCb, err)
	}

	feedd, err := feed.NewFeedDaemon(feed.MakePrivateSessionProvider(apiCli),
		feed.MakePublicSessionProvider(apiCli),
		feedCb,
	)

	if err != nil {
		panic(err)
	}
	feedCb.AddSubscription(&feedmodel.FeedCmd{Cmd: "subscribe", Args: map[string]interface{}{"t": "price", "i": "101", "m": 11}})
	log.Printf("We have feedd: %+v", feedd)

	res, err := apiCli.FeedSub("depth", "46", "11")
	log.Printf("Res::: %+v -- %+v", res, err)
	/*
	     // FOR TEST....
	     rchan := remote.MakeRequestReplyChannel(pubsub, *topic)
	     rtrans := transports.NewRemoteTransportClient(rchan)
	     cli := api.NewApiClient(rtrans)
	     acc,err := cli.Accounts()
	     log.Printf("Accounts: %+v, %+v",acc,err)


	       // --------------  Old tests
	   	l, _ := net.Listen("tcp", *bind)
	   	srv := gorpc.NewRpcTransportServer(l, ph)
	   	defer srv.Close()

	   	// Start feed.  (badly)
	   	cli := api.NewApiClient(ph)

	   	nsqips := []string{"127.0.0.1:6150"}
	   	nsqConfig := map[string]interface{}{}
	   	nf, err := nsqfeeder.NewNsqfeeder("nodent.feed", "nordnet.admin", nsqips, nsqConfig, cli)
	   	log.Printf("NSQ: %+v, err: %+v", nf, err)
	*/
	c := make(chan interface{})
	log.Print(<-c)
}
