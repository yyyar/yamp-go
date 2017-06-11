//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/format"
	"net"
	"sync"
	"testing"
)

//
// Test tcp transport
//
func TestTcp(t *testing.T) {

	const (
		PROTOCOL = "tcp"
		ADDR     = "localhost:5555"
	)

	var wg sync.WaitGroup
	wg.Add(1)

	//
	// Server code
	//
	doServer := func() {

		l, err := net.Listen(PROTOCOL, ADDR)
		if err != nil {
			t.Fatal(err)
		}

		go (func() {
			c, err := l.Accept()
			if err != nil {
				t.Log(err)
			}

			conn, _ := NewConnection(false, c, &format.JsonBodyFormat{})
			conn.OnRequest("echo", func(req *api.Request, res *api.Response) {

				var body string
				req.ReadTo(&body)

				t.Log(req.Id(), "OnRequest\t'echo': ", body)
				res.Done(body)
			})
		})()
	}

	//
	// Client code
	//
	doClient := func() {

		c, err := net.Dial(PROTOCOL, ADDR)
		if err != nil {
			t.Fatal(err)
		}

		conn, _ := NewConnection(true, c, &format.JsonBodyFormat{})
		conn.SendRequest("echo", "hello world!", func(res *api.Response) {

			var body string
			res.ReadTo(&body)

			t.Log(res.RequestId(), "OnResponse\t'echo': ", body)
			wg.Done()
		})

	}

	doServer()
	doClient()

	// Wait until everything is done
	wg.Wait()
}
