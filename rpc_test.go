package reflect_rpc_demo

import (
	"sync"
	"testing"
)

var testServerOnce sync.Once
var testClientOnce sync.Once

var testServer *Server
var testClient *Client

func newTestServer() *Server {

	f := func() {
		testServer = NewServer("tcp", "127.0.0.1:9000")
		go testServer.Start()
	}
	testServerOnce.Do(f)

	return testServer
}

func newTestClient() *Client {

	f := func() {
		testClient = NewClient("tcp", "127.0.0.1:9000", 10)
	}

	testClientOnce.Do(f)
	return testClient
}

func test_Rpc1(id int) (int, string, error) {
	return id * 10, "abc", nil
}

func TestRpc1(t *testing.T) {
	s := newTestServer()
	s.Register("rpc1", test_Rpc1)

	c := newTestClient()
	var r func(int) (int, string, error)
	if err := c.MakeRpc("rpc1", &r); err != nil {
		t.Fatal(err)
	}

	a, b, e := r(10)
	if e != nil {
		t.Fatal(e)
	}

	if a != 100 || b != "abc" {
		t.Fatal(a, b)
	}

}
