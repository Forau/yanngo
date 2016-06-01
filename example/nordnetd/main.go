package main

import (
	//	"github.com/Forau/yanngo/api"
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

	// If daemon also should run somce commands, can do it like this
	/*
	 cli := api.NewApiClient(ph)
	 res, err := cli.Accounts()
	*/

	l, _ := net.Listen("tcp", *bind)
	srv := gorpc.NewRpcTransportServer(l, ph)
	defer srv.Close()

	c := make(chan interface{})
	log.Print(<-c)
}
