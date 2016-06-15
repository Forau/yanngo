package main

import (
	"github.com/Forau/yanngo/api"
	"github.com/Forau/yanngo/feed/nsqfeeder"
	"github.com/Forau/yanngo/transports"
	"github.com/Forau/yanngo/transports/gorpc"
	"io/ioutil"
	"os"

	"flag"
	"log"

	"net"
)

var (
	user     = flag.String("user", "", "User name")
	pass     = flag.String("pass", "", "Password")
	bind     = flag.String("bind", ":2008", "Address to bind to for rcp server. Example, :1234, to bind to tcp port 1234, or : to bind to random port.  127.0.0.1:2008 would make it only accessable from local host")
	endpoint = flag.String("url", "https://api.test.nordnet.se/next/2", "The base URL.")
	pemFile  = flag.String("pem", "../../NEXTAPI_TEST_public.pem", "The PEM file")
)

func main() {
	flag.Parse()

	file, err := os.Open(*pemFile)
	if err != nil {
		panic(err)
	}
	pem, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	ph, err := transports.NewDefaultTransport(*endpoint, []byte(*user), []byte(*pass), pem)
	if err != nil {
		panic(err)
	}

	l, _ := net.Listen("tcp", *bind)
	srv := gorpc.NewRpcTransportServer(l, ph)
	defer srv.Close()

	// Start feed.  (badly)
	cli := api.NewApiClient(ph)

	nsqips := []string{"127.0.0.1:6150"}
	nsqConfig := map[string]interface{}{}
	nf, err := nsqfeeder.NewNsqfeeder("nodent.feed", "nordnet.admin", nsqips, nsqConfig, cli)
	log.Printf("NSQ: %+v, err: %+v", nf, err)

	c := make(chan interface{})
	log.Print(<-c)
}
