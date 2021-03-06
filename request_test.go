//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"fmt"
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/format"
	"io"
	"sync"
	"testing"
)

//
// MockConnection useful for testing with io.Pipe()
//
type MockConnection struct {
	io.ReadCloser
	io.Writer
}

//
// Test basic request-response operation
//
func TestReqRes(t *testing.T) {

	const N = 10

	var wg sync.WaitGroup
	wg.Add(N)

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	// Run responder
	go (func() {

		client, _ := NewConnection(true, &MockConnection{r1, w2}, &format.JsonBodyFormat{})

		client.OnRequest("sum", func(req *api.Request, res *api.Response) {

			var body []int
			(&format.JsonBodyFormat{}).Parse(req.RawBody(), &body)

			t.Log(res.RequestId(), "OnRequest   'sum': ", body)

			res.Done(body[0] + body[1])
		})

	})()

	// Run requester
	go (func() {

		server, _ := NewConnection(false, &MockConnection{r2, w1}, &format.JsonBodyFormat{})

		for i := 0; i < N; i++ {

			server.SendRequest("sum", []int{i, i}, func(res *api.Response) {

				var body int
				res.Read(&body)

				t.Log(res.RequestId(), "OnResponse  'sum': ", body)
				wg.Done()
			})

		}

	})()

	wg.Wait()
}

//
// Test progressive responses
//
func TestProgressive(t *testing.T) {

	const N = 5

	var wg sync.WaitGroup
	wg.Add(N)

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	// Run responder
	go (func() {

		client, _ := NewConnection(true, &MockConnection{r1, w2}, &format.JsonBodyFormat{})

		client.OnRequest("foo", func(req *api.Request, res *api.Response) {

			t.Log(res.RequestId(), "OnRequest   'foo'")

			for i := 0; i < N-1; i++ {
				res.Progress(fmt.Sprintf("hello %d", i))
			}

			res.Done("end")
		})

	})()

	// Run requester
	go (func() {

		server, _ := NewConnection(false, &MockConnection{r2, w1}, &format.JsonBodyFormat{})

		server.SendRequest("foo", nil, func(res *api.Response) {

			var body string
			res.Read(&body)

			t.Log("OnResponse  'foo' progress = ", res.IsProgress(), body)
			wg.Done()
		})

	})()

	wg.Wait()
}
