package feed_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/Forau/yanngo/feed"

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
	l, err := tls.Listen("tcp", "127.0.0.1:", TLS)
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
	t *testing.T
}

func (c *simpleCallback) OnConnect(w feed.CmdWriter, ft feed.FeedType) {
	c.t.Logf("Connect[%v]: %+v", ft, w)
}
func (c *simpleCallback) OnMessage(msg *feed.FeedMsg, ft feed.FeedType) {
	c.t.Logf("Msg[%v]: %+v", ft, msg.String())
}
func (c *simpleCallback) OnError(err error, ft feed.FeedType) { c.t.Logf("Err[%v]: %+v", ft, err) }

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

	cb := &simpleCallback{t}

	fd, err := feed.NewFeedDaemon(privSess, pubSess, cb)
	if err != nil {
		t.Fatalf("Daemon error: %+v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ugly yield

	err = fd.Subscribe("price", "46", "11")
	if err != nil {
		t.Fatalf("Subscribe error: %+v", err)
	}
	fmt.Println("Closing: ", fd.Close())
	close(quit)
}
