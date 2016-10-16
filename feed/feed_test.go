package feed_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"github.com/Forau/yanngo/feed"
	"github.com/Forau/yanngo/feed/feedmodel"
	"github.com/Forau/yanngo/remote"

	"math/big"
	"net"
	"testing"
	"time"
)

// We generate a self signed cert, just for this test run.
var TLS *tls.Config

func init() {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1812),
		Subject: pkix.Name{
			Country:            []string{"Internet"},
			Organization:       []string{"The Internet"},
			OrganizationalUnit: []string{"Moving bits"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{8, 8, 8, 8, 8},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		panic("create ca failed: " + err.Error())
	}

	//  certPair,_ := tls.X509KeyPair(ca_b,x509.MarshalPKCS1PrivateKey(priv))
	certPair := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	// This is the server cert. Self signed, and not very valid, so client MUST use InsecureSkipVerify
	TLS = &tls.Config{
		Certificates: []tls.Certificate{certPair},
		RootCAs:      pool,
	}

	// To make the feed accept self signed certs
	feed.DefaultTLS = &tls.Config{InsecureSkipVerify: true}
}

type testSrv struct {
	listen   net.Listener
	t        *testing.T
	exit     chan interface{}
	isClosed bool

	connFn func(c net.Conn)
}

func (ts *testSrv) Close() error {
	close(ts.exit)
	for i := 0; i < 10; i++ {
		if ts.isClosed {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("Server did not close within one second...")
}

func (ts *testSrv) mainLoop() {
	defer ts.listen.Close()
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				ts.t.Log(err)
				return
			}
			go ts.connFn(conn)
		}
	}(ts.listen)
	<-ts.exit // Wait for exit
	ts.isClosed = true
	ts.t.Logf("Exiting main loop for %+v", ts)
}

func newTestSrv(t *testing.T, connFn func(net.Conn)) (srv *testSrv) {
	// &tls.Config{RootCAs: TLS.RootCAs}
	l, err := tls.Listen("tcp", "127.0.0.1:0", TLS)
	if err != nil {
		t.Fatal(err)
	}
	srv = &testSrv{
		listen: l,
		t:      t,
		exit:   make(chan interface{}),
		connFn: connFn,
	}
	go srv.mainLoop()
	t.Logf("New test server: %+v, listening on %s", srv, srv.listen.Addr().String())
	return
}

type simpleCallback struct {
	t           *testing.T
	connectChan chan bool
}

func (c *simpleCallback) OnConnect(w feed.CmdWriter, ft feedmodel.FeedType) {
	c.t.Logf("Connect[%v]: %+v", ft, w)
	c.connectChan <- true
}
func (c *simpleCallback) OnMessage(msg *feedmodel.FeedMsg, ft feedmodel.FeedType) {
	c.t.Logf("Msg[%v]: %+v", ft, msg.String())
}
func (c *simpleCallback) OnError(err error, ft feedmodel.FeedType) { c.t.Logf("Err[%v]: %+v", ft, err) }

func TestConnectToFeed(t *testing.T) {
	quit := make(chan interface{})
	privts := newTestSrv(t, func(c net.Conn) {
		defer c.Close()
		t.Log("Got connection: ", c)
		for {
			buff := make([]byte, 1024)
			n, err := c.Read(buff)
			t.Log("Read %d bytes(%+v): %s", n, err, string(buff))
			if err != nil {
				return
			}
		}
	})
	defer privts.Close()
	privSess := func() (key, url string, err error) {
		return "PRIV", privts.listen.Addr().String(), nil
	}

	pubts := newTestSrv(t, func(c net.Conn) {
		defer c.Close()
		t.Log("Got connection: ", c)

		for {
			buff := make([]byte, 1024)
			n, err := c.Read(buff)
			t.Log("Read %d bytes(%+v): %s", n, err, string(buff))
			if err != nil {
				return
			}
		}
	})
	defer pubts.Close()
	pubSess := func() (key, url string, err error) {
		return "PUB", pubts.listen.Addr().String(), nil
	}

	cb := &simpleCallback{t, make(chan bool)}

	fd, err := feed.NewFeedDaemon(privSess, pubSess, cb)
	if err != nil {
		t.Fatalf("Daemon error: %+v", err)
	}

	go func() {
		<-cb.connectChan
		<-cb.connectChan

		err = fd.Subscribe("price", "46", "11")
		if err != nil {
			t.Fatalf("Subscribe error: %+v", err)
		}
		fmt.Println("Closing: ", fd.Close())
		close(quit)
	}()
	select {
	case <-quit:
	case <-time.After(2000 * time.Millisecond):
		t.Error("Timeout after 2000ms")
	}
}

func TestFeedTransportState(t *testing.T) {
	var lastSent []byte
	pubChan := remote.StreamTopicChannel(func(b []byte) (err error) {
		t.Logf("Sending '%s'", string(b))
		lastSent = b
		return
	})
	ft := feed.NewFeedTransport(pubChan)

	testVerMsgs := []struct{ Msg, Ver string }{
		{`{"type":"depth","data":{"i":"101","m":11,"tick_timestamp":1466185530322,"bid2":72.75,"bid_volume2":1528194,"bid_orders2":3,"ask2":73.10,"ask_volume2":322020,"ask_orders2":1,"bid3":72.70,"bid_volume3":1369705,"bid_orders3":2,"ask3":73.15,"ask_volume3":479736,"ask_orders3":2,"bid4":72.65,"bid_volume4":1909276,"bid_orders4":3,"ask4":73.25,"ask_volume4":646411,"ask_orders4":1,"bid5":72.60,"bid_volume5":30941,"bid_orders5":1,"ask5":73.35,"ask_volume5":636401,"ask_orders5":1}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185530378,"tick_timestamp":1466185530378,"last":73.05,"last_volume":204062,"turnover":284685127906.65,"turnover_volume":3900386024}}`, ""},
		{`{"type":"depth","data":{"i":"46","m":11,"tick_timestamp":1466185561373,"bid_volume2":18732,"bid_orders2":1}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561424,"last":72.90,"last_volume":96563,"turnover":285924418711.15,"turnover_volume":3917373452}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561424,"last":72.85,"last_volume":260879,"turnover":285943423746.30,"turnover_volume":3917634331}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561424,"last_volume":443968,"turnover":285975766815.10,"turnover_volume":3918078299}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561424,"last":72.90,"turnover":285924418711.15,"turnover_volume":3917373452}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561424,"turnover":285975766815.10,"turnover_volume":3918078299}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561424,"last":72.85,"turnover":285943423746.30,"turnover_volume":3917634331}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561424,"tick_timestamp":1466185561639,"bid":72.85,"bid_volume":410206}}`, ""},
		{`{"type":"depth","data":{"i":"101","m":11,"tick_timestamp":1466185561639,"bid1":72.85,"bid_volume1":410206,"bid2":72.75,"bid_volume2":911663,"bid_orders2":1,"bid3":72.70,"bid_volume3":1060440,"bid_orders3":3,"bid4":72.65,"bid_volume4":492362,"bid_orders4":1,"bid5":72.60,"bid_volume5":2743477,"bid_orders5":4}}`, ""},
		{`{"type":"price","data":{"i":"46","m":11,"trade_timestamp":1466185559346,"tick_timestamp":1466185561590,"ask":105.60,"ask_volume":700553}}`, ""},
		{`{"type":"depth","data":{"i":"46","m":11,"tick_timestamp":1466185561590,"ask1":105.60,"ask_volume1":700553,"ask_orders1":1,"ask2":105.80,"ask_volume2":705762,"ask_orders2":2,"ask3":106.00,"ask_volume3":1388331,"ask_orders3":3,"ask4":106.10,"ask_volume4":219837,"ask5":106.20,"ask_volume5":451261}}`, ""},
		{`{"type":"depth","data":{"i":"101","m":11,"tick_timestamp":1466185561669}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561935,"tick_timestamp":1466185561935,"last":73.20,"last_volume":2587,"turnover":2859759561835,"turnover_volume":3918080886}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561935,"tick_timestamp":1466185561935,"bid":72.85,"bid_volume":410206,"ask":73.20,"ask_volume":1569692,"close":60.10,"high":80.00,"last":73.20,"last_volume":2587,"low":62.30,"open":80.00,"vwap":72.98,"turnover":2859759561835,"turnover_volume":3918080886}}`, ""},
		{`{"type":"price","data":{"i":"101","m":11,"trade_timestamp":1466185561935,"tick_timestamp":1466185561944,"ask_volume":1567105}}`, `{"type":"price","data":{"ask":73.20,"ask_volume":1567105,"bid":72.85,"bid_volume":410206,"close":60.10,"high":80.00,"i":"101","last":73.20,"last_volume":2587,"low":62.30,"m":11,"open":80.00,"tick_timestamp":1466185561944,"trade_timestamp":1466185561935,"turnover":2859759561835,"turnover_volume":3918080886,"vwap":72.98}}`},
		{`{"type":"depth","data":{"i":"101","m":11,"tick_timestamp":1466185561944,"ask_volume1":1567105,"bid_volume3":1026338,"bid_orde,rs3":2,"bid4":72.60,"bid_volume4":2448007,"bid_orders4":3,"bid5":72.50,"bid_volume5":1181703,"bid_orders5":2}}`, `{"type":"depth","data":{"ask2":73.10,"ask3":73.15,"ask4":73.25,"ask5":73.35,"ask_orders2":1,"ask_orders3":2,"ask_orders4":1,"ask_orders5":1,"ask_volume1":1567105,"ask_volume2":322020,"ask_volume3":479736,"ask_volume4":646411,"ask_volume5":636401,"bid1":72.85,"bid2":72.75,"bid3":72.70,"bid4":72.60,"bid5":72.50,"bid_orde,rs3":2,"bid_orders2":1,"bid_orders3":3,"bid_orders4":3,"bid_orders5":2,"bid_volume1":410206,"bid_volume2":911663,"bid_volume3":1026338,"bid_volume4":2448007,"bid_volume5":1181703,"i":"101","m":11,"tick_timestamp":1466185561944}}`},
	}

	for _, msgver := range testVerMsgs {
		var feedMsg feedmodel.FeedMsg
		err := json.Unmarshal([]byte(msgver.Msg), &feedMsg)
		if err != nil {
			t.Errorf("Unable to unmarshal %s: %+v", msgver.Msg, err)
		} else {
			ft.OnMessage(&feedMsg, feedmodel.PublicFeedType)
			if msgver.Ver != "" {
				t.Logf("Verifying last with: %s", msgver.Ver)
				if msgver.Ver != string(lastSent) {
					t.Errorf("VERIFY error: '%s' != '%s'", msgver.Ver, string(lastSent))
				}
			}
		}
	}

}
