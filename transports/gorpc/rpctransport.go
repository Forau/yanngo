package gorpc

import (
	"github.com/Forau/yanngo/api"
	"net"
	"net/rpc"
	//   "net/http"
	//	"encoding/json"
	"log"
	//	"reflect"
	//	"strings"
)

type RpcTransportServer struct {
	Transport api.TransportHandler
	Listener  net.Listener
	Running   bool
}

func NewRpcTransportServer(listener net.Listener, sess api.TransportHandler) *RpcTransportServer {
	srv := &RpcTransportServer{Transport: sess, Listener: listener, Running: true}

	// TODO: Move to a struct for rpc, and one for networking to avoid having extra functions
	rpcSrv := rpc.NewServer()
	rpcSrv.Register(srv)
	go srv.acceptLoop(rpcSrv)
	return srv
}

func (rss *RpcTransportServer) acceptLoop(rpcSrv *rpc.Server) {
	for rss.Running {
		conn, err := rss.Listener.Accept()
		if err != nil || conn == nil {
			log.Printf("rpc.Serve: accept error: %+v, %+v", conn, err)
		} else {
			go rpcSrv.ServeConn(conn)
		}
	}
}

func (rss *RpcTransportServer) Close() error {
	rss.Running = false
	return rss.Listener.Close()
}

func (rss *RpcTransportServer) PerformRpc(req api.Request, res *api.Response) (err error) {
	*res = rss.Transport.Preform(&req)
	return nil
}

// Will implement Transport interface, and deligate all through rpc
type RpcTransportClient struct {
	client *rpc.Client
	addr   string
}

func NewRpcTransportClient(addr string) *RpcTransportClient {
	return &RpcTransportClient{addr: addr}
}

func (rsc *RpcTransportClient) GetClient() (*rpc.Client, error) {
	if rsc.client == nil {
		client, err := rpc.Dial("tcp", rsc.addr)
		if err != nil {
			return nil, err
		}
		rsc.client = client
	}
	return rsc.client, nil
}

func (rsc *RpcTransportClient) Close() error {
	if rsc.client != nil {
		return rsc.client.Close()
	}
	return nil
}

// Implement TransportHandler
func (rsc *RpcTransportClient) Preform(req *api.Request) (res api.Response) {
	cli, err := rsc.GetClient()
	if err != nil {
		res.Fail(-4, err.Error())
		return
	}

	err = cli.Call("RpcTransportServer.PerformRpc", req, &res)
	if err != nil {
		if err.Error() == "connection is shut down" {
			// TODO: Ugly test, fix me
			log.Printf("Trying to re-dial. %+v\n", rsc.client)
			rsc.client, err = rpc.Dial("tcp", rsc.addr)
			log.Printf("Re-dial done. %+v\n", rsc.client)
			return rsc.Preform(req)
		}
		res.Fail(-5, err.Error())
	}
	return
}
