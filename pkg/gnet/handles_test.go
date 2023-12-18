package gnet

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/kiga-hub/arc/protocols"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/errors"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
)

type testCodeServer struct {
	*gnet.EventServer
	addr       string
	multicore  bool
	async      bool
	codec      gnet.ICodec
	workerPool *goroutine.Pool
	//kafkaProducer *kafka.Producer
}

// OnInitComplete -
func (s *testCodeServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	return
}

// OnOpened ...
func (s *testCodeServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Println(" OnOpened")
	c.SetContext(c)
	out = []byte("sweetness\r\n")
	if c.LocalAddr() == nil {
		panic("nil local addr")
	}
	if c.RemoteAddr() == nil {
		panic("nil local addr")
	}
	return
}

// OnClosed ...
func (s *testCodeServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	fmt.Println("OnClosed")
	if err != nil {
		fmt.Printf("error occurred on closed, %v\n", err)
	}
	if c.Context() != c {
		panic("invalid context")
	}

	s.workerPool.Release()
	return
}

// React ...
func (s *testCodeServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Println("React")
	if s.async {
		_ = c.BufferLength()
		f := protocols.Frame{}
		err := f.Decode(frame)
		if err != nil {
			fmt.Println(err)
		} else {
			c.ShiftN(1)
			_ = s.workerPool.Submit(
				func() {
					_ = c.AsyncWrite(frame)
				})
			return
		}
	}
	out = []byte{}
	return
}

// Tick ...
func (s *testCodeServer) Tick() (delay time.Duration, action gnet.Action) {
	fmt.Println("Tick")
	go func() {
		startClient("tcp", s.addr, s.multicore, s.async)
	}()

	fmt.Printf("active connections:\n")
	delay = time.Second / 5
	return
}

func startClient(network, addr string, multicore, async bool) {
	fmt.Println("Start Client")
	c, err := net.Dial(network, addr)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	rd := bufio.NewReader(c)
	if network == "tcp" {
		msg, err := rd.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		if string(msg) != "sweetness\r\n" {
			panic("bad header")
		}
	}

}

func testServer(addr string, multicore, async bool, codec gnet.ICodec) {
	ts := &testCodeServer{
		addr:       addr,
		multicore:  multicore,
		async:      async,
		codec:      codec,
		workerPool: goroutine.Default(),
	}
	fmt.Println("testServer")
	must(gnet.Serve(ts, addr, gnet.WithMulticore(multicore), gnet.WithTCPKeepAlive(time.Second*5), gnet.WithCodec(codec), gnet.WithTicker(true)))
}

func must(err error) {
	if err != nil && err != errors.ErrUnsupportedProtocol {
		panic(err)
	}
}

// TestDefaultGnetServer ...
func TestDefaultGnetServer(t *testing.T) {
	svr := gnet.EventServer{}
	svr.OnInitComplete(gnet.Server{})
	svr.OnOpened(nil)
	svr.OnClosed(nil, nil)
	svr.React(nil, nil)
	svr.Tick()

}

// TestServe ...
func TestServe(t *testing.T) {
	codec := &protocols.Coder{}
	addr := fmt.Sprintf("http://:%d", 8972)
	t.Run("poll", func(t *testing.T) {
		t.Run("tcp", func(t *testing.T) {
			testServer(addr, true, true, codec)
		})
	})
}
